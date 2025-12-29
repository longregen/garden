-- name: ListCategoriesWithSources :many
SELECT
    c.category_id,
    c.name,
    COALESCE(
        json_agg(
            json_build_object(
                'id', s.id,
                'category_id', s.category_id,
                'source_uri', s.source_uri,
                'raw_source', s.raw_source
            )
        ) FILTER (WHERE s.id IS NOT NULL),
        '[]'::json
    ) as sources
FROM categories c
LEFT JOIN category_sources s ON c.category_id = s.category_id
GROUP BY c.category_id, c.name
ORDER BY c.name;

-- name: GetCategory :one
SELECT
    category_id,
    name
FROM categories
WHERE category_id = $1;

-- name: UpdateCategory :exec
UPDATE categories
SET name = $2
WHERE category_id = $1;

-- name: MergeCategories :exec
CALL merge_categories($1::uuid, $2::uuid);

-- name: CreateCategorySource :one
INSERT INTO category_sources (
    category_id,
    source_uri,
    raw_source
) VALUES (
    $1, $2, $3
) RETURNING id, category_id, source_uri, raw_source;

-- name: UpdateCategorySource :exec
UPDATE category_sources
SET
    source_uri = $2,
    raw_source = $3
WHERE id = $1;

-- name: DeleteCategorySource :exec
DELETE FROM category_sources
WHERE id = $1;

-- name: GetCategorySource :one
SELECT
    id,
    category_id,
    source_uri,
    raw_source
FROM category_sources
WHERE id = $1;
