package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// LogseqSyncUseCase defines the business operations for Logseq synchronization
type LogseqSyncUseCase interface {
	// Synchronize performs a full sync between Logseq repository and the database
	Synchronize(ctx context.Context) (*entity.SyncStats, error)

	// PerformHardSyncCheck compares all files in the Logseq folder with all entries in the database
	PerformHardSyncCheck(ctx context.Context) (*entity.SyncCheckResult, error)

	// ForceUpdateFileFromDB forces update of a git file with data from the database
	ForceUpdateFileFromDB(ctx context.Context, entityID uuid.UUID) error

	// ForceUpdateDBFromFile forces update of database entry with data from a git file
	ForceUpdateDBFromFile(ctx context.Context, pagePath string) (*entity.Entity, error)
}
