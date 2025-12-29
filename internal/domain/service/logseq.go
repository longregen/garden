package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

const (
	ConfigPrefix      = "logseq"
	RepoURLKey        = "logseq.repo_url"
	RepoPathKey       = "logseq.repo_path"
	LastSyncKey       = "logseq.last_sync"
	SyncEnabledKey    = "logseq.enabled"
	SSHKeyKey         = "logseq.ssh_key"
)

type logseqSyncService struct {
	configService input.ConfigurationUseCase
	entityRepo    output.EntityRepository
}

// NewLogseqSyncService creates a new Logseq sync service
func NewLogseqSyncService(
	configService input.ConfigurationUseCase,
	entityRepo output.EntityRepository,
) input.LogseqSyncUseCase {
	return &logseqSyncService{
		configService: configService,
		entityRepo:    entityRepo,
	}
}

func (s *logseqSyncService) Synchronize(ctx context.Context) (*entity.SyncStats, error) {
	// Check if sync is enabled
	syncEnabled, err := s.configService.GetBoolValue(ctx, SyncEnabledKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to check if sync is enabled: %w", err)
	}
	if !syncEnabled {
		return nil, fmt.Errorf("logseq synchronization is not enabled")
	}

	// Get repository URL and local path
	repoURL, err := s.configService.GetValue(ctx, RepoURLKey)
	if err != nil || repoURL == nil {
		return nil, fmt.Errorf("logseq repository URL is not configured")
	}

	repoPath, err := s.configService.GetValue(ctx, RepoPathKey)
	if err != nil || repoPath == nil {
		return nil, fmt.Errorf("logseq repository local path is not configured")
	}

	// Initialize stats
	stats := &entity.SyncStats{
		Errors: make([]string, 0),
	}

	// Clone or pull repository
	if err := s.ensureRepository(ctx, *repoURL, *repoPath); err != nil {
		return nil, fmt.Errorf("failed to ensure repository: %w", err)
	}

	// Get last sync timestamp
	lastSyncStr, err := s.configService.GetValue(ctx, LastSyncKey)
	var lastSync time.Time
	if err == nil && lastSyncStr != nil {
		lastSync, _ = time.Parse(time.RFC3339, *lastSyncStr)
	}

	// Sync Logseq pages to entities
	if err := s.syncPagesToEntities(ctx, *repoPath, lastSync, stats); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Error syncing pages to entities: %v", err))
	}

	// Sync entities to Logseq pages
	if err := s.syncEntitiesToPages(ctx, *repoPath, lastSync, stats); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Error syncing entities to pages: %v", err))
	}

	// Commit and push changes
	if err := s.commitAndPushChanges(ctx, *repoPath, stats); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Error committing and pushing: %v", err))
	}

	// Update last sync timestamp
	_, err = s.configService.SetConfiguration(ctx, LastSyncKey, time.Now().Format(time.RFC3339), false)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Error updating last sync time: %v", err))
	}

	return stats, nil
}

func (s *logseqSyncService) PerformHardSyncCheck(ctx context.Context) (*entity.SyncCheckResult, error) {
	repoPath, err := s.configService.GetValue(ctx, RepoPathKey)
	if err != nil || repoPath == nil {
		return nil, fmt.Errorf("logseq repository local path is not configured")
	}

	result := &entity.SyncCheckResult{
		MissingInDB:  make([]entity.Entity, 0),
		MissingInGit: make([]entity.Entity, 0),
		OutOfSync:    make([]entity.OutOfSyncItem, 0),
	}

	// Get all entities with logseq_path property
	entities, err := s.entityRepo.ListEntitiesByProperty(ctx, "logseq_path", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list entities with logseq_path: %w", err)
	}

	// Get all markdown files
	pagesDir := filepath.Join(*repoPath, "pages")
	files, err := s.getMarkdownFiles(pagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get markdown files: %w", err)
	}

	processedPaths := make(map[string]bool)

	// Check each entity against its file
	for _, ent := range entities {
		var props map[string]interface{}
		if err := json.Unmarshal(ent.Properties, &props); err != nil {
			continue
		}

		logseqPath, ok := props["logseq_path"].(string)
		if !ok || logseqPath == "" {
			continue
		}

		processedPaths[logseqPath] = true

		// Check if file exists
		if !s.fileExists(logseqPath) {
			result.MissingInGit = append(result.MissingInGit, ent)
			continue
		}

		// Parse the page
		page, err := s.parseLogseqPage(logseqPath)
		if err != nil {
			lastSyncDB := s.getLastSyncFromProps(props)
			result.OutOfSync = append(result.OutOfSync, entity.OutOfSyncItem{
				Entity:      ent,
				PagePath:    logseqPath,
				LastSyncDB:  lastSyncDB,
				LastSyncGit: nil,
			})
			continue
		}

		// Check if timestamps match
		dbLastSync := s.getLastSyncFromProps(props)
		gitLastSync := s.parseTime(page.Frontmatter.LastSync)

		if !s.timesEqual(dbLastSync, gitLastSync) || page.Frontmatter.ID != ent.EntityID.String() {
			result.OutOfSync = append(result.OutOfSync, entity.OutOfSyncItem{
				Entity:      ent,
				PagePath:    logseqPath,
				LastSyncDB:  dbLastSync,
				LastSyncGit: gitLastSync,
			})
		}
	}

	// Check for files without corresponding entities
	for _, file := range files {
		filePath := filepath.Join(pagesDir, file)
		if processedPaths[filePath] {
			continue
		}

		page, err := s.parseLogseqPage(filePath)
		if err != nil {
			continue
		}

		if page.Frontmatter.ID != "" {
			entityID, err := uuid.Parse(page.Frontmatter.ID)
			if err == nil {
				ent, err := s.entityRepo.GetEntity(ctx, entityID)
				if err == nil && ent != nil {
					lastSyncGit := s.parseTime(page.Frontmatter.LastSync)
					result.OutOfSync = append(result.OutOfSync, entity.OutOfSyncItem{
						Entity:      *ent,
						PagePath:    filePath,
						LastSyncDB:  nil,
						LastSyncGit: lastSyncGit,
					})
					continue
				}
			}
		}

		// Missing in DB
		desc := s.extractDescription(page.Content)
		propsMap := map[string]interface{}{
			"content":      page.Content,
			"logseq_path":  filePath,
		}
		propsJSON, _ := json.Marshal(propsMap)

		result.MissingInDB = append(result.MissingInDB, entity.Entity{
			EntityID:    uuid.Nil,
			Name:        page.Title,
			Type:        "note",
			Description: desc,
			Properties:  propsJSON,
			UpdatedAt:   page.LastModified,
		})
	}

	return result, nil
}

func (s *logseqSyncService) ForceUpdateFileFromDB(ctx context.Context, entityID uuid.UUID) error {
	repoPath, err := s.configService.GetValue(ctx, RepoPathKey)
	if err != nil || repoPath == nil {
		return fmt.Errorf("logseq repository local path is not configured")
	}

	ent, err := s.entityRepo.GetEntity(ctx, entityID)
	if err != nil {
		return fmt.Errorf("entity not found: %w", err)
	}

	stats := &entity.SyncStats{}
	pagePath, err := s.findOrCreatePageForEntity(*repoPath, ent, stats)
	if err != nil {
		return fmt.Errorf("failed to find or create page: %w", err)
	}

	frontmatter := entity.LogseqPageFrontmatter{
		ID:       ent.EntityID.String(),
		Title:    ent.Name,
		LastSync: time.Now().Format(time.RFC3339),
	}

	content := s.createPageContent(ent)
	if err := s.writePageFile(pagePath, frontmatter, content); err != nil {
		return fmt.Errorf("failed to write page file: %w", err)
	}

	// Update entity properties
	var props map[string]interface{}
	if err := json.Unmarshal(ent.Properties, &props); err != nil {
		props = make(map[string]interface{})
	}
	props["logseq_path"] = pagePath
	props["logseq_last_sync"] = time.Now().Format(time.RFC3339)
	propsJSON, _ := json.Marshal(props)
	rawProps := json.RawMessage(propsJSON)

	_, err = s.entityRepo.UpdateEntity(ctx, entityID, entity.UpdateEntityInput{
		Properties: &rawProps,
	})

	return err
}

func (s *logseqSyncService) ForceUpdateDBFromFile(ctx context.Context, pagePath string) (*entity.Entity, error) {
	if !s.fileExists(pagePath) {
		return nil, fmt.Errorf("file does not exist: %s", pagePath)
	}

	page, err := s.parseLogseqPage(pagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}

	var ent *entity.Entity
	if page.Frontmatter.ID != "" {
		entityID, err := uuid.Parse(page.Frontmatter.ID)
		if err == nil {
			ent, _ = s.entityRepo.GetEntity(ctx, entityID)
		}
	}

	props := map[string]interface{}{
		"content":           page.Content,
		"logseq_path":       pagePath,
		"logseq_last_sync":  time.Now().Format(time.RFC3339),
	}
	propsJSON, _ := json.Marshal(props)
	rawProps := json.RawMessage(propsJSON)

	if ent != nil {
		// Update existing
		desc := s.extractDescription(page.Content)
		name := page.Title

		updated, err := s.entityRepo.UpdateEntity(ctx, ent.EntityID, entity.UpdateEntityInput{
			Name:        &name,
			Description: desc,
			Properties:  &rawProps,
		})
		return updated, err
	}

	// Create new
	desc := s.extractDescription(page.Content)
	created, err := s.entityRepo.CreateEntity(ctx, entity.CreateEntityInput{
		Name:        page.Title,
		Type:        "note",
		Description: desc,
		Properties:  &rawProps,
	})
	if err != nil {
		return nil, err
	}

	// Update page frontmatter with new entity ID
	page.Frontmatter.ID = created.EntityID.String()
	page.Frontmatter.LastSync = time.Now().Format(time.RFC3339)
	if err := s.updatePageFrontmatter(pagePath, page.Frontmatter); err != nil {
		return created, err
	}

	return created, nil
}

// Private helper methods

func (s *logseqSyncService) ensureRepository(ctx context.Context, repoURL, repoPath string) error {
	sshKey, err := s.configService.GetValue(ctx, SSHKeyKey)
	var sshKeyPath string
	if err == nil && sshKey != nil && *sshKey != "" {
		sshKeyPath, err = s.createTempSSHKey(*sshKey)
		if err != nil {
			return err
		}
		defer os.Remove(sshKeyPath)
	}

	gitDir := filepath.Join(repoPath, ".git")
	repoExists := s.directoryExists(gitDir)

	if repoExists {
		cmd := exec.Command("git", "pull")
		cmd.Dir = repoPath
		if sshKeyPath != "" {
			cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", sshKeyPath))
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git pull failed: %w, output: %s", err, output)
		}
	} else {
		cmd := exec.Command("git", "clone", repoURL, repoPath)
		if sshKeyPath != "" {
			cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", sshKeyPath))
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git clone failed: %w, output: %s", err, output)
		}
	}

	return nil
}

func (s *logseqSyncService) syncPagesToEntities(ctx context.Context, repoPath string, lastSync time.Time, stats *entity.SyncStats) error {
	pagesDir := filepath.Join(repoPath, "pages")
	files, err := s.getMarkdownFiles(pagesDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		stats.PagesProcessed++
		filePath := filepath.Join(pagesDir, file)

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			stats.PagesSkipped++
			continue
		}

		if !lastSync.IsZero() && fileInfo.ModTime().Before(lastSync) {
			stats.PagesSkipped++
			continue
		}

		page, err := s.parseLogseqPage(filePath)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to parse %s: %v", file, err))
			stats.PagesSkipped++
			continue
		}

		if page.Frontmatter.ID != "" {
			if err := s.updateEntityFromPage(ctx, page, stats); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to update entity from %s: %v", file, err))
			}
		} else {
			if _, err := s.createEntityFromPage(ctx, page, stats); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to create entity from %s: %v", file, err))
			}
		}
	}

	return nil
}

func (s *logseqSyncService) syncEntitiesToPages(ctx context.Context, repoPath string, lastSync time.Time, stats *entity.SyncStats) error {
	var entities []entity.Entity
	var err error

	if !lastSync.IsZero() {
		entities, err = s.entityRepo.ListEntitiesUpdatedSince(ctx, lastSync)
	} else {
		entities, err = s.entityRepo.ListEntitiesUpdatedSince(ctx, time.Time{})
	}

	if err != nil {
		return err
	}

	for _, ent := range entities {
		stats.EntitiesProcessed++

		pagePath, err := s.findOrCreatePageForEntity(repoPath, &ent, stats)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to find/create page for entity %s: %v", ent.EntityID, err))
			continue
		}

		if err := s.updatePageFromEntity(ctx, pagePath, &ent, stats); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to update page from entity %s: %v", ent.EntityID, err))
		}
	}

	return nil
}

func (s *logseqSyncService) commitAndPushChanges(ctx context.Context, repoPath string, stats *entity.SyncStats) error {
	sshKey, err := s.configService.GetValue(ctx, SSHKeyKey)
	var sshKeyPath string
	if err == nil && sshKey != nil && *sshKey != "" {
		sshKeyPath, err = s.createTempSSHKey(*sshKey)
		if err != nil {
			return err
		}
		defer os.Remove(sshKeyPath)
	}

	// Check if there are changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	if len(bytes.TrimSpace(output)) == 0 {
		return nil
	}

	// Add all changes
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Commit
	timestamp := time.Now().Format(time.RFC3339)
	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Sync with system at %s", timestamp))
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	// Push
	cmd = exec.Command("git", "push")
	cmd.Dir = repoPath
	if sshKeyPath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", sshKeyPath))
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w, output: %s", err, output)
	}

	return nil
}

func (s *logseqSyncService) parseLogseqPage(filePath string) (*entity.LogseqPage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Check for Hugo templates
	contentStr := string(content)
	if strings.Contains(contentStr, "{{") && strings.Contains(contentStr, "}}") {
		if matched, _ := regexp.MatchString(`\{\{\s*\.[A-Za-z]+\s*\}\}`, contentStr); matched {
			return nil, fmt.Errorf("file appears to be a Hugo template")
		}
	}

	var frontmatter entity.LogseqPageFrontmatter
	var pageContent string

	// Parse frontmatter
	if bytes.HasPrefix(content, []byte("---\n")) || bytes.HasPrefix(content, []byte("---\r\n")) {
		parts := bytes.SplitN(content, []byte("\n---\n"), 2)
		if len(parts) < 2 {
			parts = bytes.SplitN(content, []byte("\r\n---\r\n"), 2)
		}

		if len(parts) == 2 {
			frontmatterBytes := bytes.TrimPrefix(parts[0], []byte("---\n"))
			frontmatterBytes = bytes.TrimPrefix(frontmatterBytes, []byte("---\r\n"))

			if err := yaml.Unmarshal(frontmatterBytes, &frontmatter); err != nil {
				return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
			}
			pageContent = string(bytes.TrimSpace(parts[1]))
		} else {
			pageContent = contentStr
		}
	} else {
		pageContent = contentStr
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	title := filename
	titleMatch := regexp.MustCompile(`^#\s+(.+)$`).FindStringSubmatch(pageContent)
	if len(titleMatch) > 1 {
		title = strings.TrimSpace(titleMatch[1])
	}

	return &entity.LogseqPage{
		Path:         filePath,
		Filename:     filename,
		Title:        title,
		Content:      pageContent,
		Frontmatter:  frontmatter,
		LastModified: fileInfo.ModTime(),
	}, nil
}

func (s *logseqSyncService) updateEntityFromPage(ctx context.Context, page *entity.LogseqPage, stats *entity.SyncStats) error {
	entityID, err := uuid.Parse(page.Frontmatter.ID)
	if err != nil {
		return err
	}

	ent, err := s.entityRepo.GetEntity(ctx, entityID)
	if err != nil || ent == nil {
		_, err := s.createEntityFromPage(ctx, page, stats)
		return err
	}

	var props map[string]interface{}
	if err := json.Unmarshal(ent.Properties, &props); err != nil {
		props = make(map[string]interface{})
	}
	props["content"] = page.Content
	props["logseq_path"] = page.Path
	props["logseq_last_sync"] = time.Now().Format(time.RFC3339)
	propsJSON, _ := json.Marshal(props)
	rawProps := json.RawMessage(propsJSON)

	desc := s.extractDescription(page.Content)
	name := page.Title

	_, err = s.entityRepo.UpdateEntity(ctx, entityID, entity.UpdateEntityInput{
		Name:        &name,
		Description: desc,
		Properties:  &rawProps,
	})

	if err == nil {
		stats.EntitiesUpdated++
	}

	return err
}

func (s *logseqSyncService) createEntityFromPage(ctx context.Context, page *entity.LogseqPage, stats *entity.SyncStats) (*entity.Entity, error) {
	props := map[string]interface{}{
		"content":           page.Content,
		"logseq_path":       page.Path,
		"logseq_last_sync":  time.Now().Format(time.RFC3339),
	}
	propsJSON, _ := json.Marshal(props)
	rawProps := json.RawMessage(propsJSON)

	desc := s.extractDescription(page.Content)
	ent, err := s.entityRepo.CreateEntity(ctx, entity.CreateEntityInput{
		Name:        page.Title,
		Type:        "note",
		Description: desc,
		Properties:  &rawProps,
	})

	if err != nil {
		return nil, err
	}

	stats.EntitiesCreated++

	// Update page frontmatter
	page.Frontmatter.ID = ent.EntityID.String()
	page.Frontmatter.LastSync = time.Now().Format(time.RFC3339)
	if err := s.updatePageFrontmatter(page.Path, page.Frontmatter); err != nil {
		return ent, err
	}

	return ent, nil
}

func (s *logseqSyncService) findOrCreatePageForEntity(repoPath string, ent *entity.Entity, stats *entity.SyncStats) (string, error) {
	var props map[string]interface{}
	if err := json.Unmarshal(ent.Properties, &props); err == nil {
		if logseqPath, ok := props["logseq_path"].(string); ok && logseqPath != "" {
			if s.fileExists(logseqPath) {
				return logseqPath, nil
			}
		}
	}

	pagesDir := filepath.Join(repoPath, "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		return "", err
	}

	filename := s.sanitizeFilename(ent.Name) + ".md"
	pagePath := filepath.Join(pagesDir, filename)

	frontmatter := entity.LogseqPageFrontmatter{
		ID:       ent.EntityID.String(),
		Title:    ent.Name,
		LastSync: time.Now().Format(time.RFC3339),
	}

	content := s.createPageContent(ent)
	if err := s.writePageFile(pagePath, frontmatter, content); err != nil {
		return "", err
	}

	stats.PagesCreated++
	return pagePath, nil
}

func (s *logseqSyncService) updatePageFromEntity(ctx context.Context, pagePath string, ent *entity.Entity, stats *entity.SyncStats) error {
	page, err := s.parseLogseqPage(pagePath)
	if err != nil {
		// Can't parse, create new page
		repoPath := filepath.Dir(filepath.Dir(pagePath))
		newPagePath, err := s.findOrCreatePageForEntity(repoPath, ent, stats)
		if err != nil {
			return err
		}

		if newPagePath != pagePath {
			var props map[string]interface{}
			if err := json.Unmarshal(ent.Properties, &props); err != nil {
				props = make(map[string]interface{})
			}
			props["logseq_path"] = newPagePath
			props["logseq_last_sync"] = time.Now().Format(time.RFC3339)
			propsJSON, _ := json.Marshal(props)
			rawProps := json.RawMessage(propsJSON)

			_, err = s.entityRepo.UpdateEntity(ctx, ent.EntityID, entity.UpdateEntityInput{
				Properties: &rawProps,
			})
		}
		return err
	}

	lastSync := s.parseTime(page.Frontmatter.LastSync)

	if lastSync != nil && page.LastModified.After(*lastSync) && ent.UpdatedAt.After(*lastSync) {
		if ent.UpdatedAt.After(page.LastModified) {
			return s.updatePageContent(pagePath, ent, page.Frontmatter)
		}
		stats.PagesSkipped++
		return nil
	}

	if lastSync == nil || ent.UpdatedAt.After(*lastSync) {
		if err := s.updatePageContent(pagePath, ent, page.Frontmatter); err != nil {
			return err
		}
		stats.PagesUpdated++
		return nil
	}

	stats.PagesSkipped++
	return nil
}

func (s *logseqSyncService) updatePageContent(pagePath string, ent *entity.Entity, existingFrontmatter entity.LogseqPageFrontmatter) error {
	frontmatter := existingFrontmatter
	frontmatter.ID = ent.EntityID.String()
	frontmatter.Title = ent.Name
	frontmatter.LastSync = time.Now().Format(time.RFC3339)

	content := s.createPageContent(ent)
	return s.writePageFile(pagePath, frontmatter, content)
}

func (s *logseqSyncService) createPageContent(ent *entity.Entity) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n\n", ent.Name))

	if ent.Description != nil && *ent.Description != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", *ent.Description))
	}

	var props map[string]interface{}
	if err := json.Unmarshal(ent.Properties, &props); err == nil {
		if pageContent, ok := props["content"].(string); ok && pageContent != "" {
			// Remove existing title to avoid duplication
			re := regexp.MustCompile(`^#\s+.+$`)
			contentWithoutTitle := re.ReplaceAllString(pageContent, "")
			contentWithoutTitle = strings.TrimSpace(contentWithoutTitle)
			if contentWithoutTitle != "" {
				content.WriteString(fmt.Sprintf("%s\n\n", contentWithoutTitle))
			}
		}
	}

	return strings.TrimSpace(content.String())
}

func (s *logseqSyncService) extractDescription(content string) *string {
	re := regexp.MustCompile(`^#\s+.+$`)
	contentWithoutTitle := re.ReplaceAllString(content, "")
	contentWithoutTitle = strings.TrimSpace(contentWithoutTitle)

	paragraphs := strings.Split(contentWithoutTitle, "\n\n")
	if len(paragraphs) > 0 && strings.TrimSpace(paragraphs[0]) != "" {
		desc := strings.TrimSpace(paragraphs[0])
		return &desc
	}

	return nil
}

func (s *logseqSyncService) updatePageFrontmatter(pagePath string, frontmatter entity.LogseqPageFrontmatter) error {
	content, err := os.ReadFile(pagePath)
	if err != nil {
		return err
	}

	contentStr := string(content)
	var pageContent string

	if bytes.HasPrefix(content, []byte("---\n")) || bytes.HasPrefix(content, []byte("---\r\n")) {
		parts := bytes.SplitN(content, []byte("\n---\n"), 2)
		if len(parts) < 2 {
			parts = bytes.SplitN(content, []byte("\r\n---\r\n"), 2)
		}
		if len(parts) == 2 {
			pageContent = string(bytes.TrimSpace(parts[1]))
		} else {
			pageContent = contentStr
		}
	} else {
		pageContent = contentStr
	}

	return s.writePageFile(pagePath, frontmatter, pageContent)
}

func (s *logseqSyncService) writePageFile(pagePath string, frontmatter entity.LogseqPageFrontmatter, content string) error {
	dir := filepath.Dir(pagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")

	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return err
	}
	buf.Write(yamlBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(content)

	return os.WriteFile(pagePath, buf.Bytes(), 0644)
}

func (s *logseqSyncService) getMarkdownFiles(dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func (s *logseqSyncService) fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (s *logseqSyncService) directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (s *logseqSyncService) sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = re.ReplaceAllString(name, "_")
	name = strings.ReplaceAll(name, " ", "_")
	return strings.ToLower(name)
}

func (s *logseqSyncService) createTempSSHKey(key string) (string, error) {
	tmpFile, err := os.CreateTemp("", "logseq_ssh_key_*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(key); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func (s *logseqSyncService) getLastSyncFromProps(props map[string]interface{}) *time.Time {
	if lastSyncStr, ok := props["logseq_last_sync"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastSyncStr); err == nil {
			return &t
		}
	}
	return nil
}

func (s *logseqSyncService) parseTime(timeStr string) *time.Time {
	if timeStr == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return &t
	}
	return nil
}

func (s *logseqSyncService) timesEqual(t1, t2 *time.Time) bool {
	if t1 == nil && t2 == nil {
		return true
	}
	if t1 == nil || t2 == nil {
		return false
	}
	return t1.Unix() == t2.Unix()
}
