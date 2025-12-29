package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

// ContactService implements the ContactUseCase interface
type ContactService struct {
	repo output.ContactRepository
}

// NewContactService creates a new contact service
func NewContactService(repo output.ContactRepository) *ContactService {
	return &ContactService{
		repo: repo,
	}
}

func (s *ContactService) GetContact(ctx context.Context, contactID uuid.UUID) (*entity.FullContact, error) {
	contact, err := s.repo.GetContact(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	evaluation, err := s.repo.GetContactEvaluation(ctx, contactID)
	if err != nil {
		// Evaluation might not exist, that's ok
		evaluation = nil
	}

	tags, err := s.repo.GetContactTags(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact tags: %w", err)
	}

	knownNames, err := s.repo.GetContactKnownNames(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get known names: %w", err)
	}

	rooms, err := s.repo.GetContactRooms(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}

	sources, err := s.repo.GetContactSources(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}

	return &entity.FullContact{
		Contact:    *contact,
		Evaluation: evaluation,
		Tags:       tags,
		KnownNames: knownNames,
		Rooms:      rooms,
		Sources:    sources,
	}, nil
}

func (s *ContactService) ListContacts(ctx context.Context, page, pageSize int32, searchQuery *string) (*input.PaginatedResponse[entity.Contact], error) {
	offset := (page - 1) * pageSize

	var contacts []entity.Contact
	var err error

	if searchQuery != nil && *searchQuery != "" {
		pattern := "%" + *searchQuery + "%"
		contacts, err = s.repo.SearchContacts(ctx, pattern, pageSize, offset)
	} else {
		contacts, err = s.repo.ListContacts(ctx, pageSize, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	total, err := s.repo.CountContacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count contacts: %w", err)
	}

	// Get contact IDs for batch loading
	contactIDs := make([]uuid.UUID, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ContactID
	}

	// Load evaluations and tags in batch
	evaluations, err := s.repo.GetEvaluationsByContactIDs(ctx, contactIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get evaluations: %w", err)
	}

	tags, err := s.repo.GetTagsByContactIDs(ctx, contactIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	// Merge evaluation data into contacts
	for i := range contacts {
		if eval, ok := evaluations[contacts[i].ContactID]; ok {
			// We can't directly add these to Contact entity, so they'll be handled in the handler/DTO layer
			_ = eval
		}
		if contactTags, ok := tags[contacts[i].ContactID]; ok {
			_ = contactTags
		}
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &input.PaginatedResponse[entity.Contact]{
		Data:       contacts,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ContactService) CreateContact(ctx context.Context, input entity.CreateContactInput) (*entity.Contact, error) {
	contact, err := s.repo.CreateContact(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}
	return contact, nil
}

func (s *ContactService) UpdateContact(ctx context.Context, contactID uuid.UUID, input entity.UpdateContactInput) error {
	err := s.repo.UpdateContact(ctx, contactID, input)
	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}
	return nil
}

func (s *ContactService) DeleteContact(ctx context.Context, contactID uuid.UUID) error {
	err := s.repo.DeleteContact(ctx, contactID)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	return nil
}

func (s *ContactService) UpdateEvaluation(ctx context.Context, contactID uuid.UUID, input entity.UpdateEvaluationInput) error {
	exists, err := s.repo.EvaluationExists(ctx, contactID)
	if err != nil {
		return fmt.Errorf("failed to check evaluation existence: %w", err)
	}

	if exists {
		err = s.repo.UpdateEvaluation(ctx, contactID, input)
	} else {
		err = s.repo.CreateEvaluation(ctx, contactID, input)
	}

	if err != nil {
		return fmt.Errorf("failed to update evaluation: %w", err)
	}
	return nil
}

func (s *ContactService) MergeContacts(ctx context.Context, sourceID, targetID uuid.UUID) error {
	err := s.repo.MergeContacts(ctx, sourceID, targetID)
	if err != nil {
		return fmt.Errorf("failed to merge contacts: %w", err)
	}
	return nil
}

func (s *ContactService) BatchUpdate(ctx context.Context, updates []entity.BatchContactUpdate) ([]entity.BatchUpdateResult, error) {
	results := make([]entity.BatchUpdateResult, 0, len(updates))

	for _, update := range updates {
		result := entity.BatchUpdateResult{
			ContactID: update.ContactID,
			Success:   false,
		}

		// Update basic contact info if name or notes is provided
		if update.Name != nil || update.Notes != nil {
			updateInput := entity.UpdateContactInput{
				Name:  update.Name,
				Notes: update.Notes,
			}
			if err := s.repo.UpdateContact(ctx, update.ContactID, updateInput); err != nil {
				errMsg := err.Error()
				result.Error = &errMsg
				results = append(results, result)
				continue
			}
		}

		// Update evaluations if provided
		if update.Importance != nil || update.Closeness != nil || update.Fondness != nil {
			evalInput := entity.UpdateEvaluationInput{
				Importance: update.Importance,
				Closeness:  update.Closeness,
				Fondness:   update.Fondness,
			}
			if err := s.UpdateEvaluation(ctx, update.ContactID, evalInput); err != nil {
				errMsg := err.Error()
				result.Error = &errMsg
				results = append(results, result)
				continue
			}
		}

		// Handle tags if provided
		if update.Tags != nil {
			currentTags, err := s.repo.GetContactTags(ctx, update.ContactID)
			if err != nil {
				errMsg := err.Error()
				result.Error = &errMsg
				results = append(results, result)
				continue
			}

			// Find tags to add
			currentTagMap := make(map[uuid.UUID]bool)
			for _, tag := range currentTags {
				currentTagMap[tag.TagID] = true
			}

			updateTagMap := make(map[uuid.UUID]bool)
			for _, tag := range update.Tags {
				updateTagMap[tag.TagID] = true
				if !currentTagMap[tag.TagID] {
					// Add new tag
					if err := s.repo.AddContactTag(ctx, update.ContactID, tag.TagID); err != nil {
						errMsg := err.Error()
						result.Error = &errMsg
						results = append(results, result)
						continue
					}
				}
			}

			// Find tags to remove
			for _, tag := range currentTags {
				if !updateTagMap[tag.TagID] {
					if err := s.repo.RemoveContactTag(ctx, update.ContactID, tag.TagID); err != nil {
						errMsg := err.Error()
						result.Error = &errMsg
						results = append(results, result)
						continue
					}
				}
			}
		}

		result.Success = true
		results = append(results, result)
	}

	return results, nil
}

func (s *ContactService) RefreshStats(ctx context.Context) error {
	if err := s.repo.EnsureContactStatsExist(ctx); err != nil {
		return fmt.Errorf("failed to ensure stats exist: %w", err)
	}

	if err := s.repo.UpdateContactMessageStats(ctx); err != nil {
		return fmt.Errorf("failed to update message stats: %w", err)
	}

	return nil
}

func (s *ContactService) GetContactTags(ctx context.Context, contactID uuid.UUID) ([]entity.ContactTag, error) {
	tags, err := s.repo.GetContactTags(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact tags: %w", err)
	}
	return tags, nil
}

func (s *ContactService) AddTag(ctx context.Context, contactID uuid.UUID, tagName string) (*entity.ContactTag, error) {
	// Check if tag exists
	tag, err := s.repo.GetContactTagByName(ctx, tagName)
	if err != nil {
		// Tag doesn't exist, create it
		tag, err = s.repo.CreateTagName(ctx, tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to create tag: %w", err)
		}
	}

	// Check if contact already has this tag
	exists, err := s.repo.ContactTagExists(ctx, contactID, tag.TagID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tag existence: %w", err)
	}

	// Add tag to contact if it doesn't exist
	if !exists {
		if err := s.repo.AddContactTag(ctx, contactID, tag.TagID); err != nil {
			return nil, fmt.Errorf("failed to add tag to contact: %w", err)
		}
	}

	return &entity.ContactTag{
		TagID: tag.TagID,
		Name:  tag.Name,
	}, nil
}

func (s *ContactService) RemoveTag(ctx context.Context, contactID uuid.UUID, tagID uuid.UUID) error {
	err := s.repo.RemoveContactTag(ctx, contactID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}
	return nil
}

func (s *ContactService) ListAllTags(ctx context.Context) ([]entity.ContactTagName, error) {
	tags, err := s.repo.GetAllTagNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tags: %w", err)
	}
	return tags, nil
}

func (s *ContactService) ListAllContactSources(ctx context.Context) ([]entity.ContactSource, error) {
	sources, err := s.repo.ListAllContactSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list contact sources: %w", err)
	}
	return sources, nil
}

func (s *ContactService) CreateContactSource(ctx context.Context, input entity.CreateContactSourceInput) (*entity.ContactSource, error) {
	source, err := s.repo.CreateContactSource(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact source: %w", err)
	}
	return source, nil
}

func (s *ContactService) UpdateContactSource(ctx context.Context, id uuid.UUID, input entity.UpdateContactSourceInput) (*entity.ContactSource, error) {
	source, err := s.repo.UpdateContactSource(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact source: %w", err)
	}
	return source, nil
}

func (s *ContactService) DeleteContactSource(ctx context.Context, id uuid.UUID) error {
	err := s.repo.DeleteContactSource(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact source: %w", err)
	}
	return nil
}
