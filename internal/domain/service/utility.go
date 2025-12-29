package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

var startTime = time.Now()

// UtilityService implements the utility use case
type UtilityService struct {
	sessionRepo output.SessionRepository
	messageRepo output.MessageRepository
	configRepo  output.ConfigurationRepository
	pool        *pgxpool.Pool
}

// NewUtilityService creates a new utility service
func NewUtilityService(
	sessionRepo output.SessionRepository,
	messageRepo output.MessageRepository,
	configRepo output.ConfigurationRepository,
	pool *pgxpool.Pool,
) *UtilityService {
	return &UtilityService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		configRepo:  configRepo,
		pool:        pool,
	}
}

// GetDebugInfo returns system debug information
func (s *UtilityService) GetDebugInfo(ctx context.Context) (*entity.DebugInfo, error) {
	// Check database connection
	dbStatus := "connected"
	if err := s.pool.Ping(ctx); err != nil {
		dbStatus = "disconnected: " + err.Error()
	}

	// Get non-secret configuration values
	configs, err := s.configRepo.ListConfigurations(ctx, entity.ConfigurationFilter{})
	configMap := make(map[string]string)
	if err == nil {
		sensitiveKeywords := []string{"secret", "password", "token", "key", "credential", "auth"}
		for _, cfg := range configs {
			isSensitive := false
			keyLower := strings.ToLower(cfg.Key)
			for _, keyword := range sensitiveKeywords {
				if strings.Contains(keyLower, keyword) {
					isSensitive = true
					break
				}
			}
			if !isSensitive {
				configMap[cfg.Key] = cfg.Value
			}
		}
	}

	// Calculate uptime
	uptime := time.Since(startTime).Round(time.Second).String()

	return &entity.DebugInfo{
		DatabaseStatus: dbStatus,
		Version:        "1.0.0",
		Uptime:         uptime,
		Config:         configMap,
	}, nil
}

// CleanupStaleConversations removes sessions older than 30 days with no messages
func (s *UtilityService) CleanupStaleConversations(ctx context.Context) (int64, error) {
	olderThan := time.Now().AddDate(0, 0, -30)
	return s.sessionRepo.DeleteStaleConversations(ctx, olderThan)
}

// GetMessagesByIDs retrieves messages by their IDs
func (s *UtilityService) GetMessagesByIDs(ctx context.Context, messageIDs []uuid.UUID) ([]entity.Message, error) {
	return s.messageRepo.GetMessagesByIDs(ctx, messageIDs)
}
