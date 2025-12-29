package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
	"github.com/google/uuid"
)

// NoteService implements the NoteUseCase interface
type NoteService struct {
	repo              output.NoteRepository
	embeddingsService output.EmbeddingsService
}

// NewNoteService creates a new note service
func NewNoteService(repo output.NoteRepository, embeddingsService output.EmbeddingsService) *NoteService {
	return &NoteService{
		repo:              repo,
		embeddingsService: embeddingsService,
	}
}

var (
	// Matches [[text]] or [[display][text]]
	entityRefRegex = regexp.MustCompile(`\[\[(.*?)(?:\]\[([^\]]+))?\]\]`)
	uuidRegex      = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// processContentForStorage converts entity references to entity IDs
// Converts:
// - [[Entity Name]] to [[entity-id]]
// - [[Display Text][Entity Name]] to [[Display Text][entity-id]]
func (s *NoteService) processContentForStorage(ctx context.Context, content string, sourceType string, sourceID uuid.UUID) (string, error) {
	if content == "" {
		return content, nil
	}

	processedContent := content
	matches := entityRefRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		original := match[0]
		firstPart := match[1]
		secondPart := ""
		if len(match) > 2 {
			secondPart = match[2]
		}

		var entityID uuid.UUID
		var displayText string

		if secondPart != "" {
			// Format: [[Display Text][Entity Name or ID]]
			if uuidRegex.MatchString(secondPart) {
				// Second part is already a UUID
				parsedUUID, err := uuid.Parse(secondPart)
				if err != nil {
					continue
				}
				entityID = parsedUUID
				displayText = firstPart

				// Verify the entity exists
				_, err = s.repo.GetEntityByID(ctx, entityID)
				if err != nil {
					// Entity doesn't exist, skip this reference
					continue
				}
			} else {
				// Second part is an entity name
				existingEntity, err := s.repo.GetEntityByName(ctx, secondPart)
				if err != nil || existingEntity == nil {
					// Create new entity
					createdEntityID, err := s.repo.CreateEntity(ctx, secondPart, "general", nil, make(map[string]interface{}))
					if err != nil {
						continue
					}
					entityID = *createdEntityID
				} else {
					entityID = existingEntity.EntityID
				}
				displayText = firstPart
			}

			// Replace with [[Display Text][entity-id]]
			replacement := fmt.Sprintf("[[%s][%s]]", displayText, entityID.String())
			processedContent = strings.Replace(processedContent, original, replacement, 1)
		} else {
			// Format: [[Entity Name]]
			existingEntity, err := s.repo.GetEntityByName(ctx, firstPart)
			if err != nil || existingEntity == nil {
				// Create new entity
				createdEntityID, err := s.repo.CreateEntity(ctx, firstPart, "general", nil, make(map[string]interface{}))
				if err != nil {
					continue
				}
				entityID = *createdEntityID
			} else {
				entityID = existingEntity.EntityID
			}
			displayText = firstPart

			// Replace with [[entity-id]]
			replacement := fmt.Sprintf("[[%s]]", entityID.String())
			processedContent = strings.Replace(processedContent, original, replacement, 1)
		}

		// Record the reference
		ref := entity.EntityReference{
			SourceType:    sourceType,
			SourceID:      sourceID,
			EntityID:      entityID,
			ReferenceText: displayText,
			Position:      nil,
		}
		if err := s.repo.CreateEntityReference(ctx, ref); err != nil {
			// Log error but continue
			fmt.Printf("Error creating entity reference: %v\n", err)
		}
	}

	return processedContent, nil
}

// processContentForDisplay converts entity IDs to display text with markdown links
// Converts:
// - [[entity-id]] to [Entity Name](/entities/entity-id)
// - [[Display Text][entity-id]] to [Display Text](/entities/entity-id)
func (s *NoteService) processContentForDisplay(ctx context.Context, content string) (string, error) {
	if content == "" {
		return content, nil
	}

	processedContent := content
	matches := entityRefRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		original := match[0]
		firstPart := match[1]
		secondPart := ""
		if len(match) > 2 {
			secondPart = match[2]
		}

		if secondPart != "" {
			// Format: [[display][entity-id]]
			entityID, err := uuid.Parse(secondPart)
			if err != nil {
				continue
			}

			entity, err := s.repo.GetEntityByID(ctx, entityID)
			if err == nil && entity != nil {
				replacement := fmt.Sprintf("[%s](/entities/%s)", firstPart, secondPart)
				processedContent = strings.Replace(processedContent, original, replacement, 1)
			}
		} else {
			// Format: [[entity-id]]
			entityID, err := uuid.Parse(firstPart)
			if err != nil {
				continue
			}

			entity, err := s.repo.GetEntityByID(ctx, entityID)
			if err == nil && entity != nil {
				replacement := fmt.Sprintf("[%s](/entities/%s)", entity.Name, firstPart)
				processedContent = strings.Replace(processedContent, original, replacement, 1)
			}
		}
	}

	return processedContent, nil
}

func (s *NoteService) GetNote(ctx context.Context, noteID uuid.UUID) (*entity.FullNote, error) {
	note, tags, err := s.repo.GetNoteWithTags(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Get entity ID for this note
	entityID, err := s.repo.GetNoteEntityRelationship(ctx, noteID)
	if err != nil {
		// Entity relationship might not exist, that's ok
		entityID = nil
	}

	// Process content for display
	var processedContent *string
	if note.Contents != nil {
		processed, err := s.processContentForDisplay(ctx, *note.Contents)
		if err != nil {
			return nil, fmt.Errorf("failed to process content: %w", err)
		}
		processedContent = &processed
	}

	return &entity.FullNote{
		Note:             *note,
		Tags:             tags,
		EntityID:         entityID,
		ProcessedContent: processedContent,
	}, nil
}

func (s *NoteService) ListNotes(ctx context.Context, page, pageSize int32, searchQuery *string) (*input.PaginatedResponse[entity.NoteListItem], error) {
	offset := (page - 1) * pageSize

	var notes []entity.NoteListItem
	var total int64
	var err error

	if searchQuery != nil && *searchQuery != "" {
		pattern := "%" + *searchQuery + "%"
		notes, err = s.repo.SearchNotes(ctx, pattern, pageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to search notes: %w", err)
		}
		total, err = s.repo.CountSearchNotes(ctx, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to count search results: %w", err)
		}
	} else {
		notes, err = s.repo.ListNotes(ctx, pageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list notes: %w", err)
		}
		total, err = s.repo.CountNotes(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count notes: %w", err)
		}
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &input.PaginatedResponse[entity.NoteListItem]{
		Data:       notes,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *NoteService) CreateNote(ctx context.Context, input entity.CreateNoteInput) (*entity.FullNote, error) {
	// Create the note first to get a real ID
	note, err := s.repo.CreateNote(ctx, input.Title, "", input.Contents)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// Process content and create entity references
	processedContent := input.Contents
	if input.Contents != "" {
		processedContent, err = s.processContentForStorage(ctx, input.Contents, "note", note.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to process content: %w", err)
		}

		// Update note with processed content
		if err := s.repo.UpdateNote(ctx, note.ID, input.Title, processedContent); err != nil {
			return nil, fmt.Errorf("failed to update note content: %w", err)
		}
		note.Contents = &processedContent
	}

	// Create note entity
	entityID, err := s.repo.CreateEntity(ctx, input.Title, "note", nil, make(map[string]interface{}))
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	// Create entity relationship
	if err := s.repo.CreateEntityRelationship(ctx, *entityID, "item", note.ID, "identity", make(map[string]interface{})); err != nil {
		return nil, fmt.Errorf("failed to create entity relationship: %w", err)
	}

	// Handle tags
	currentTimestamp := note.Created
	if len(input.Tags) > 0 {
		for _, tagName := range input.Tags {
			tag, err := s.repo.UpsertNoteTag(ctx, tagName, currentTimestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to upsert tag: %w", err)
			}

			if err := s.repo.CreateItemTag(ctx, note.ID, tag.ID); err != nil {
				return nil, fmt.Errorf("failed to create item tag: %w", err)
			}
		}
	}

	// Return the created note
	return s.GetNote(ctx, note.ID)
}

func (s *NoteService) UpdateNote(ctx context.Context, noteID uuid.UUID, input entity.UpdateNoteInput) (*entity.FullNote, error) {
	// Get existing note
	note, err := s.repo.GetNote(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Prepare update values
	title := ""
	if input.Title != nil {
		title = *input.Title
	} else if note.Title != nil {
		title = *note.Title
	}

	contents := ""
	if input.Contents != nil {
		contents = *input.Contents

		// Delete existing entity references
		if err := s.repo.DeleteNoteEntityReferences(ctx, noteID); err != nil {
			return nil, fmt.Errorf("failed to delete entity references: %w", err)
		}

		// Process content and create new entity references
		processedContent, err := s.processContentForStorage(ctx, contents, "note", noteID)
		if err != nil {
			return nil, fmt.Errorf("failed to process content: %w", err)
		}
		contents = processedContent
	} else if note.Contents != nil {
		contents = *note.Contents
	}

	// Update the note
	if err := s.repo.UpdateNote(ctx, noteID, title, contents); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	// Handle tags if provided
	if input.Tags != nil {
		// Delete existing tags
		if err := s.repo.DeleteNoteItemTags(ctx, noteID); err != nil {
			return nil, fmt.Errorf("failed to delete tags: %w", err)
		}

		// Insert new tags
		currentTimestamp := note.Modified
		for _, tagName := range *input.Tags {
			tag, err := s.repo.UpsertNoteTag(ctx, tagName, currentTimestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to upsert tag: %w", err)
			}

			if err := s.repo.CreateItemTag(ctx, noteID, tag.ID); err != nil {
				return nil, fmt.Errorf("failed to create item tag: %w", err)
			}
		}
	}

	return s.GetNote(ctx, noteID)
}

func (s *NoteService) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	// Delete item tags
	if err := s.repo.DeleteNoteItemTags(ctx, noteID); err != nil {
		return fmt.Errorf("failed to delete item tags: %w", err)
	}

	// Delete semantic index
	if err := s.repo.DeleteNoteSemanticIndex(ctx, noteID); err != nil {
		// Ignore error if semantic index doesn't exist
	}

	// Delete entity references
	if err := s.repo.DeleteNoteEntityReferences(ctx, noteID); err != nil {
		return fmt.Errorf("failed to delete entity references: %w", err)
	}

	// Get entity ID for this note
	entityID, err := s.repo.GetNoteEntityRelationship(ctx, noteID)
	if err == nil && entityID != nil {
		// Soft delete the entity
		if err := s.repo.SoftDeleteEntity(ctx, *entityID); err != nil {
			return fmt.Errorf("failed to soft delete entity: %w", err)
		}

		// Delete entity relationships
		if err := s.repo.DeleteEntityRelationships(ctx, *entityID); err != nil {
			return fmt.Errorf("failed to delete entity relationships: %w", err)
		}
	}

	// Delete the note
	if err := s.repo.DeleteNote(ctx, noteID); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

func (s *NoteService) ListAllTags(ctx context.Context) ([]entity.NoteTag, error) {
	tags, err := s.repo.GetAllNoteTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tags: %w", err)
	}
	return tags, nil
}

func (s *NoteService) SearchSimilarNotes(ctx context.Context, query string, strategy string) ([]entity.NoteListItem, error) {
	if strategy == "" {
		strategy = "qa-v2-passage"
	}

	embeddings, err := s.embeddingsService.GetEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	results, err := s.repo.SearchSimilarNotes(ctx, embeddings[0].Embedding, strategy, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar notes: %w", err)
	}

	return results, nil
}
