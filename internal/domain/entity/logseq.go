package entity

import (
	"time"

	"github.com/google/uuid"
)

// LogseqPageFrontmatter represents the YAML frontmatter in a Logseq page
type LogseqPageFrontmatter struct {
	ID       string            `yaml:"id,omitempty"`
	Title    string            `yaml:"title,omitempty"`
	LastSync string            `yaml:"last_sync,omitempty"`
	Extra    map[string]interface{} `yaml:",inline"`
}

// LogseqPage represents a parsed Logseq markdown page
type LogseqPage struct {
	Path         string
	Filename     string
	Title        string
	Content      string
	Frontmatter  LogseqPageFrontmatter
	LastModified time.Time
}

// SyncStats represents statistics from a sync operation
type SyncStats struct {
	PagesProcessed     int
	PagesCreated       int
	PagesUpdated       int
	PagesSkipped       int
	EntitiesProcessed  int
	EntitiesCreated    int
	EntitiesUpdated    int
	EntitiesSkipped    int
	Errors             []string
}

// SyncCheckResult represents the result of a hard sync check
type SyncCheckResult struct {
	MissingInDB  []Entity
	MissingInGit []Entity
	OutOfSync    []OutOfSyncItem
}

// OutOfSyncItem represents an entity that is out of sync between DB and Git
type OutOfSyncItem struct {
	Entity       Entity
	PagePath     string
	LastSyncDB   *time.Time
	LastSyncGit  *time.Time
}

// ForceUpdateRequest represents a request to force update from DB to Git
type ForceUpdateRequest struct {
	EntityID uuid.UUID `json:"entity_id"`
}
