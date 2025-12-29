package input

import (
	"context"

	"garden3/internal/domain/entity"
)

// DashboardUseCase defines the business operations for dashboard statistics
type DashboardUseCase interface {
	// GetStats retrieves comprehensive dashboard statistics
	GetStats(ctx context.Context) (*entity.DashboardStats, error)
}
