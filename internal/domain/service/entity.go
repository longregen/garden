package service

import (
	"context"
	"fmt"
	"regexp"

	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
	"github.com/google/uuid"
)

// EntityService implements the EntityUseCase interface
type EntityService struct {
	repo output.EntityRepository
}

// NewEntityService creates a new entity service
func NewEntityService(repo output.EntityRepository) *EntityService {
	return &EntityService{
		repo: repo,
	}
}

func (s *EntityService) GetEntity(ctx context.Context, entityID uuid.UUID) (*entity.Entity, error) {
	ent, err := s.repo.GetEntity(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}
	return ent, nil
}

func (s *EntityService) UpdateEntity(ctx context.Context, entityID uuid.UUID, input entity.UpdateEntityInput) (*entity.Entity, error) {
	// Update the entity
	updatedEntity, err := s.repo.UpdateEntity(ctx, entityID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// If name was updated and entity type is 'person', update related contacts
	if input.Name != nil && updatedEntity.Type == "person" {
		// Get relationships where related_type='contact' and relationship_type='identity'
		err = s.repo.UpdateContactNameByRelationship(ctx, entityID, "identity", *input.Name)
		if err != nil {
			// Log error but don't fail the update
			// In production, you'd use proper logging
			fmt.Printf("warning: failed to update related contacts: %v\n", err)
		}
	}

	return updatedEntity, nil
}

func (s *EntityService) DeleteEntity(ctx context.Context, entityID uuid.UUID) error {
	if err := s.repo.DeleteEntity(ctx, entityID); err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	return nil
}

func (s *EntityService) GetEntityRelationships(ctx context.Context, entityID uuid.UUID, relatedType, relationshipType *string) ([]entity.EntityRelationship, error) {
	// If both filters are provided, use the optimized query
	if relatedType != nil && relationshipType != nil {
		rels, err := s.repo.GetEntityRelationshipsByTypeAndRelationship(ctx, entityID, *relatedType, *relationshipType)
		if err != nil {
			return nil, fmt.Errorf("failed to get entity relationships: %w", err)
		}
		return rels, nil
	}

	// Otherwise get all and filter in memory if needed
	allRels, err := s.repo.GetEntityRelationships(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity relationships: %w", err)
	}

	// Apply filters if provided
	if relatedType == nil && relationshipType == nil {
		return allRels, nil
	}

	var filtered []entity.EntityRelationship
	for _, rel := range allRels {
		if relatedType != nil && rel.RelatedType != *relatedType {
			continue
		}
		if relationshipType != nil && rel.RelationshipType != *relationshipType {
			continue
		}
		filtered = append(filtered, rel)
	}

	return filtered, nil
}

func (s *EntityService) CreateEntityRelationship(ctx context.Context, input entity.CreateEntityRelationshipInput) (*entity.EntityRelationship, error) {
	var metadata []byte
	if input.Metadata != nil {
		metadata = []byte(input.Metadata)
	}

	rel, err := s.repo.CreateEntityRelationship(ctx, input.EntityID, input.RelatedType, input.RelatedID, input.RelationshipType, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity relationship: %w", err)
	}
	return rel, nil
}

func (s *EntityService) DeleteEntityRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	if err := s.repo.DeleteEntityRelationship(ctx, relationshipID); err != nil {
		return fmt.Errorf("failed to delete entity relationship: %w", err)
	}
	return nil
}

func (s *EntityService) GetEntityReferences(ctx context.Context, entityID uuid.UUID, sourceType *string) ([]entity.EntityReference, error) {
	refs, err := s.repo.GetEntityReferences(ctx, entityID, sourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity references: %w", err)
	}
	return refs, nil
}

func (s *EntityService) ParseEntityReferences(ctx context.Context, content string) ([]entity.ParsedReference, error) {
	if content == "" {
		return []entity.ParsedReference{}, nil
	}

	// Regex: \[\[(.*?)(?:\]\[([^\]]+))?\]\]
	// Match 1: entity name
	// Match 2 (optional): display text
	const pattern = `\[\[(.*?)(?:\]\[([^\]]+))?\]\]`
	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringSubmatch(content, -1)
	references := make([]entity.ParsedReference, 0, len(matches))

	for _, match := range matches {
		original := match[0]
		entityName := match[1]
		displayText := entityName

		if len(match) > 2 && match[2] != "" {
			displayText = match[2]
		}

		references = append(references, entity.ParsedReference{
			Original:    original,
			EntityName:  entityName,
			DisplayText: displayText,
		})
	}

	return references, nil
}

func (s *EntityService) ListEntities(ctx context.Context, entityType string, updatedSince *string) ([]entity.Entity, error) {
	entities, err := s.repo.ListEntities(ctx, entityType, updatedSince)
	if err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}
	return entities, nil
}

func (s *EntityService) CreateEntity(ctx context.Context, input entity.CreateEntityInput) (*entity.Entity, error) {
	created, err := s.repo.CreateEntity(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}
	return created, nil
}

func (s *EntityService) SearchEntities(ctx context.Context, query string, entityType string) ([]entity.Entity, error) {
	entities, err := s.repo.SearchEntities(ctx, query, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}
	return entities, nil
}

func (s *EntityService) ListDeletedEntities(ctx context.Context) ([]entity.Entity, error) {
	entities, err := s.repo.ListDeletedEntities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list deleted entities: %w", err)
	}
	return entities, nil
}
