-- name: GetTagByName :one
SELECT id, name, created, modified, last_activity
FROM tags
WHERE name = $1
LIMIT 1;

-- name: UpsertTagRecord :one
INSERT INTO tags (
    name,
    created,
    modified,
    last_activity
) VALUES (
    $1,
    $2,
    $2,
    $2
)
ON CONFLICT (name) DO UPDATE SET
    modified = $2,
    last_activity = $2
RETURNING id, name, created, modified, last_activity;

-- name: CreateItemTagRelationRecord :exec
INSERT INTO item_tags (item_id, tag_id)
VALUES ($1, $2)
ON CONFLICT (item_id, tag_id) DO NOTHING;

-- name: DeleteItemTagRelationRecord :exec
DELETE FROM item_tags
WHERE item_id = $1 AND tag_id = $2;

-- name: ListAllTagsRecords :many
SELECT id, name, created, modified, last_activity
FROM tags
ORDER BY name ASC;

-- name: ListAllTagsWithUsageRecords :many
SELECT
    t.id,
    t.name,
    t.created,
    t.modified,
    t.last_activity,
    COUNT(it.item_id) as usage_count
FROM tags t
LEFT JOIN item_tags it ON t.id = it.tag_id
GROUP BY t.id, t.name, t.created, t.modified, t.last_activity
ORDER BY usage_count DESC, t.name ASC;

-- name: GetItemsByTagID :many
SELECT
    i.id,
    i.title,
    i.created,
    i.modified,
    array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
FROM items i
INNER JOIN item_tags it ON i.id = it.item_id
LEFT JOIN item_tags it2 ON i.id = it2.item_id
LEFT JOIN tags t ON it2.tag_id = t.id
WHERE it.tag_id = $1
GROUP BY i.id, i.title, i.created, i.modified
ORDER BY i.modified DESC
LIMIT $2 OFFSET $3;

-- name: CountItemsByTagID :one
SELECT COUNT(DISTINCT i.id)
FROM items i
INNER JOIN item_tags it ON i.id = it.item_id
WHERE it.tag_id = $1;

-- name: DeleteTagRecord :exec
DELETE FROM tags WHERE id = $1;

-- name: DeleteAllTagRelationsRecord :exec
DELETE FROM item_tags WHERE tag_id = $1;
