-- name: ListEntities :many
SELECT
    entity_id,
    name,
    type,
    description,
    properties,
    created_at,
    updated_at,
    deleted_at
FROM entities
WHERE deleted_at IS NULL
  AND ($1 = '' OR type = $1)
  AND ($2::timestamp IS NULL OR updated_at >= $2)
ORDER BY name;

-- name: GetEntity :one
SELECT
    entity_id,
    name,
    type,
    description,
    properties,
    created_at,
    updated_at,
    deleted_at
FROM entities
WHERE entity_id = $1 AND deleted_at IS NULL;

-- name: CreateEntity :one
INSERT INTO entities (
    name,
    type,
    description,
    properties,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    NOW()
)
RETURNING entity_id, name, type, description, properties, created_at, updated_at, deleted_at;

-- name: UpdateEntity :one
UPDATE entities
SET
    name = COALESCE($2, name),
    type = COALESCE($3, type),
    description = COALESCE($4, description),
    properties = COALESCE($5, properties),
    updated_at = NOW()
WHERE entity_id = $1 AND deleted_at IS NULL
RETURNING entity_id, name, type, description, properties, created_at, updated_at, deleted_at;

-- name: SearchEntities :many
SELECT
    entity_id,
    name,
    type,
    description,
    properties,
    created_at,
    updated_at,
    deleted_at
FROM entities
WHERE deleted_at IS NULL
  AND name ILIKE '%' || $1 || '%'
  AND ($2 = '' OR type = $2)
ORDER BY name;

-- name: ListDeletedEntities :many
SELECT
    entity_id,
    name,
    type,
    description,
    properties,
    created_at,
    updated_at,
    deleted_at
FROM entities
WHERE deleted_at IS NOT NULL
ORDER BY deleted_at DESC;

-- name: DeleteEntity :exec
UPDATE entities
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE entity_id = $1;

-- name: GetEntityRelationshipsByEntity :many
SELECT
    id,
    entity_id,
    related_type,
    related_id,
    relationship_type,
    metadata,
    created_at,
    updated_at
FROM entity_relationships
WHERE entity_id = $1;

-- name: GetEntityRelationshipsByTypeAndRelationship :many
SELECT
    id,
    entity_id,
    related_type,
    related_id,
    relationship_type,
    metadata,
    created_at,
    updated_at
FROM entity_relationships
WHERE entity_id = $1
  AND related_type = $2
  AND relationship_type = $3;

-- name: UpdateContactNameByRelationship :exec
UPDATE contacts
SET
    name = $3,
    last_update = NOW()
WHERE contact_id IN (
    SELECT related_id
    FROM entity_relationships
    WHERE entity_id = $1
      AND related_type = 'contact'
      AND relationship_type = $2
);

-- name: CreateEntityRelationship :one
INSERT INTO entity_relationships (
    entity_id,
    related_type,
    related_id,
    relationship_type,
    metadata
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, entity_id, related_type, related_id, relationship_type, metadata, created_at, updated_at;

-- name: DeleteEntityRelationship :exec
DELETE FROM entity_relationships
WHERE id = $1;

-- name: GetEntityReferences :many
SELECT
    id,
    source_type,
    source_id,
    entity_id,
    reference_text,
    position,
    created_at
FROM entity_references
WHERE entity_id = $1;

-- name: GetEntityReferencesBySourceType :many
SELECT
    id,
    source_type,
    source_id,
    entity_id,
    reference_text,
    position,
    created_at
FROM entity_references
WHERE entity_id = $1
  AND source_type = $2;
