-- name: GetNote :one
SELECT
    i.id,
    i.title,
    i.slug,
    i.contents,
    i.created,
    i.modified
FROM items i
WHERE i.id = $1;

-- name: GetNoteWithTags :one
SELECT
    i.id,
    i.title,
    i.slug,
    i.contents,
    i.created,
    i.modified,
    array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
FROM items i
LEFT JOIN item_tags it ON i.id = it.item_id
LEFT JOIN tags t ON it.tag_id = t.id
WHERE i.id = $1
GROUP BY i.id, i.title, i.slug, i.contents, i.created, i.modified;

-- name: GetNoteEntityRelationship :one
SELECT entity_id
FROM entity_relationships
WHERE related_type = 'item'
  AND related_id = $1
  AND relationship_type = 'identity'
LIMIT 1;

-- name: ListNotes :many
SELECT
    i.id,
    i.title,
    i.created,
    i.modified,
    array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
FROM items i
LEFT JOIN item_tags it ON i.id = it.item_id
LEFT JOIN tags t ON it.tag_id = t.id
GROUP BY i.id, i.title, i.created, i.modified
ORDER BY i.modified DESC
LIMIT $1 OFFSET $2;

-- name: SearchNotes :many
SELECT
    i.id,
    i.title,
    i.created,
    i.modified,
    array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
FROM items i
LEFT JOIN item_tags it ON i.id = it.item_id
LEFT JOIN tags t ON it.tag_id = t.id
WHERE i.title ILIKE $1 OR i.contents ILIKE $1
GROUP BY i.id, i.title, i.created, i.modified
ORDER BY i.modified DESC
LIMIT $2 OFFSET $3;

-- name: CountNotes :one
SELECT COUNT(*) FROM items;

-- name: CountSearchNotes :one
SELECT COUNT(*) FROM items
WHERE title ILIKE $1 OR contents ILIKE $1;

-- name: CreateNote :one
INSERT INTO items (
    title,
    slug,
    contents,
    created,
    modified
) VALUES (
    $1, $2, $3,
    EXTRACT(epoch FROM CURRENT_TIMESTAMP),
    EXTRACT(epoch FROM CURRENT_TIMESTAMP)
) RETURNING id, title, slug, contents, created, modified;

-- name: UpdateNote :exec
UPDATE items
SET
    title = $2,
    contents = $3,
    modified = EXTRACT(epoch FROM CURRENT_TIMESTAMP)
WHERE id = $1;

-- name: DeleteNote :exec
DELETE FROM items WHERE id = $1;

-- name: DeleteNoteItemTags :exec
DELETE FROM item_tags WHERE item_id = $1;

-- name: DeleteNoteSemanticIndex :exec
DELETE FROM item_semantic_index WHERE item_id = $1;

-- name: DeleteNoteEntityReferences :exec
DELETE FROM entity_references
WHERE source_type = 'note' AND source_id = $1;

-- name: GetNoteEntityRelationshipForDelete :one
SELECT entity_id
FROM entity_relationships
WHERE related_type = 'item'
  AND related_id = $1
  AND relationship_type = 'identity'
LIMIT 1;

-- name: SoftDeleteEntity :exec
UPDATE entities
SET
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE entity_id = $1;

-- name: DeleteEntityRelationships :exec
DELETE FROM entity_relationships
WHERE entity_id = $1;

-- name: CreateEntityForNote :one
INSERT INTO entities (
    name,
    type,
    description,
    properties,
    updated_at
) VALUES (
    $1, $2, $3, $4, CURRENT_TIMESTAMP
) RETURNING entity_id, name, type, description, properties, created_at, updated_at;

-- name: CreateEntityRelationshipForNote :exec
INSERT INTO entity_relationships (
    entity_id,
    related_type,
    related_id,
    relationship_type,
    metadata,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, CURRENT_TIMESTAMP
);

-- name: CreateEntityReference :exec
INSERT INTO entity_references (
    source_type,
    source_id,
    entity_id,
    reference_text,
    position
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetEntityByID :one
SELECT entity_id, name, type, description, properties, created_at, updated_at, deleted_at
FROM entities
WHERE entity_id = $1;

-- name: GetEntityByName :one
SELECT entity_id, name, type, description, properties, created_at, updated_at, deleted_at
FROM entities
WHERE name = $1
  AND deleted_at IS NULL
LIMIT 1;

-- Tags operations (specific to items/notes)
-- name: GetNoteTagByName :one
SELECT id, name, created, modified, last_activity
FROM tags
WHERE name = $1
LIMIT 1;

-- name: UpsertNoteTag :one
INSERT INTO tags (
    name,
    created,
    modified,
    last_activity
) VALUES (
    $1,
    EXTRACT(epoch FROM CURRENT_TIMESTAMP),
    EXTRACT(epoch FROM CURRENT_TIMESTAMP),
    $2
)
ON CONFLICT (name) DO UPDATE
SET
    last_activity = $2,
    modified = EXTRACT(epoch FROM CURRENT_TIMESTAMP)
RETURNING id, name, created, modified, last_activity;

-- name: CreateItemTag :exec
INSERT INTO item_tags (item_id, tag_id)
VALUES ($1, $2);

-- name: GetAllNoteTags :many
SELECT id, name, last_activity, created, modified
FROM tags
ORDER BY last_activity DESC NULLS LAST;

-- name: SearchSimilarNotes :many
SELECT
    i.id,
    i.title,
    i.created,
    i.modified,
    array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
FROM items i
INNER JOIN item_semantic_index isi ON i.id = isi.item_id
LEFT JOIN item_tags it ON i.id = it.item_id
LEFT JOIN tags t ON it.tag_id = t.id
GROUP BY i.id, i.title, i.created, i.modified, isi.embedding
ORDER BY isi.embedding <=> $1::vector
LIMIT $2;
