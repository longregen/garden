-- name: GetItemByID :one
SELECT
    i.id,
    i.title,
    i.slug,
    i.contents,
    i.created,
    i.modified
FROM items i
WHERE i.id = $1;

-- name: GetItemWithTagsByID :one
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

-- name: ListAllItems :many
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

-- name: CountAllItems :one
SELECT COUNT(*) FROM items;

-- name: CreateItemRecord :one
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

-- name: UpdateItemRecord :exec
UPDATE items
SET
    title = $2,
    contents = $3,
    modified = EXTRACT(epoch FROM CURRENT_TIMESTAMP)
WHERE id = $1;

-- name: DeleteItemRecord :exec
DELETE FROM items WHERE id = $1;

-- name: DeleteItemTagsRelation :exec
DELETE FROM item_tags WHERE item_id = $1;

-- name: DeleteItemSemanticIndexRecord :exec
DELETE FROM item_semantic_index WHERE item_id = $1;

-- name: GetItemTagByTagName :one
SELECT id, name, created, modified, last_activity
FROM tags
WHERE name = $1
LIMIT 1;

-- name: UpsertItemTagRecord :one
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

-- name: CreateItemTagLinkage :exec
INSERT INTO item_tags (item_id, tag_id)
VALUES ($1, $2);

-- name: GetItemTagsList :one
SELECT
    array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
FROM items i
LEFT JOIN item_tags it ON i.id = it.item_id
LEFT JOIN tags t ON it.tag_id = t.id
WHERE i.id = $1
GROUP BY i.id;

-- name: SearchSimilarItemsByEmbedding :many
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
