package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

// SessionService implements the session use cases
type SessionService struct {
	repo             output.SessionRepository
	embeddingService output.EmbeddingService
}

// NewSessionService creates a new session service
func NewSessionService(repo output.SessionRepository, embeddingService output.EmbeddingService) input.SessionUseCase {
	return &SessionService{
		repo:             repo,
		embeddingService: embeddingService,
	}
}

// SearchSessions performs semantic search using embeddings, falls back to text search
func (s *SessionService) SearchSessions(ctx context.Context, query string, limit int32) ([]entity.SessionSearchResult, error) {
	// Try embedding search first
	embedding, err := s.embeddingService.GetEmbedding(ctx, query)
	if err == nil && len(embedding) > 0 {
		results, err := s.repo.SearchSessionsWithEmbeddings(ctx, embedding, limit)
		if err == nil && len(results) > 0 {
			return results, nil
		}
	}

	// Fall back to text search
	searchPattern := "%" + query + "%"
	return s.repo.SearchSessionsWithText(ctx, searchPattern, limit)
}

// GetRoomSessions retrieves all session summaries for a room
func (s *SessionService) GetRoomSessions(ctx context.Context, roomID uuid.UUID) ([]entity.SessionSummary, error) {
	return s.repo.GetSessionSummaries(ctx, roomID)
}

// SearchContactSessions searches sessions where a contact participated
func (s *SessionService) SearchContactSessions(ctx context.Context, contactID uuid.UUID, searchTerm string) ([]entity.SessionSearchResult, error) {
	searchPattern := "%" + searchTerm + "%"
	return s.repo.SearchContactSessionSummaries(ctx, contactID, searchPattern)
}

// GetTimeline retrieves timeline visualization data for a room
func (s *SessionService) GetTimeline(ctx context.Context, roomID uuid.UUID) (*entity.TimelineData, error) {
	// Get sessions with message counts
	sessions, err := s.repo.GetSessionsForRoom(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Get participant activity
	participantActivity, err := s.repo.GetSessionParticipantActivity(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant activity: %w", err)
	}

	// Build participant map by session
	participantsBySession := make(map[uuid.UUID][]entity.TimelineParticipant)
	for _, pa := range participantActivity {
		participantsBySession[pa.SessionID] = append(participantsBySession[pa.SessionID], entity.TimelineParticipant{
			ID:           pa.SenderContactID,
			Name:         pa.SenderName,
			MessageCount: pa.MessageCount,
		})
	}

	// Group sessions by month
	sessionsByMonth := make(map[string][]entity.TimelineSession)
	for _, session := range sessions {
		monthKey := session.FirstDateTime.Format("2006-01")

		// Calculate duration in minutes
		var duration int64
		if session.LastDateTime != nil {
			duration = int64(session.LastDateTime.Sub(session.FirstDateTime).Minutes())
		}

		timelineSession := entity.TimelineSession{
			SessionID:     session.SessionID,
			FirstDateTime: session.FirstDateTime,
			LastDateTime:  session.LastDateTime,
			MessageCount:  session.MessageCount,
			Duration:      duration,
			Participants:  participantsBySession[session.SessionID],
		}

		sessionsByMonth[monthKey] = append(sessionsByMonth[monthKey], timelineSession)
	}

	// Convert to array and calculate aggregates
	var timelineData []entity.TimelineMonth
	for month, monthlySessions := range sessionsByMonth {
		var totalMessages int64
		var totalDuration int64
		for _, session := range monthlySessions {
			totalMessages += session.MessageCount
			totalDuration += session.Duration
		}

		avgDuration := float64(0)
		if len(monthlySessions) > 0 {
			avgDuration = float64(totalDuration) / float64(len(monthlySessions))
		}

		timelineData = append(timelineData, entity.TimelineMonth{
			Month:           month,
			Sessions:        monthlySessions,
			TotalSessions:   len(monthlySessions),
			TotalMessages:   totalMessages,
			AverageDuration: avgDuration,
		})
	}

	// Sort by month
	sort.Slice(timelineData, func(i, j int) bool {
		return timelineData[i].Month < timelineData[j].Month
	})

	// Calculate overall stats
	var firstSessionDate, lastSessionDate *time.Time
	if len(sessions) > 0 {
		firstSessionDate = &sessions[0].FirstDateTime
		if sessions[len(sessions)-1].LastDateTime != nil {
			lastSessionDate = sessions[len(sessions)-1].LastDateTime
		}
	}

	return &entity.TimelineData{
		TimelineData:     timelineData,
		TotalSessions:    len(sessions),
		FirstSessionDate: firstSessionDate,
		LastSessionDate:  lastSessionDate,
	}, nil
}

// GetSessionMessages retrieves all messages in a session with contact info
func (s *SessionService) GetSessionMessages(ctx context.Context, sessionID uuid.UUID) (*entity.SessionMessagesResponse, error) {
	messages, err := s.repo.GetSessionMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}

	// Extract unique contact IDs
	contactIDSet := make(map[uuid.UUID]bool)
	for _, msg := range messages {
		contactIDSet[msg.SenderContactID] = true
	}

	contactIDs := make([]uuid.UUID, 0, len(contactIDSet))
	for id := range contactIDSet {
		contactIDs = append(contactIDs, id)
	}

	// Fetch contacts
	var contactsMap map[uuid.UUID]entity.SessionMessageContact
	if len(contactIDs) > 0 {
		contacts, err := s.repo.GetContactsByIDs(ctx, contactIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get contacts: %w", err)
		}

		contactsMap = make(map[uuid.UUID]entity.SessionMessageContact)
		for _, contact := range contacts {
			contactsMap[contact.ContactID] = contact
		}
	}

	return &entity.SessionMessagesResponse{
		Messages: messages,
		Contacts: contactsMap,
	}, nil
}
