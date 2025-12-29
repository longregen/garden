-- name: SearchAll :many
WITH search_union AS (
    -- Contacts
    SELECT
        'contact'::text as item_type,
        contact_id::text as item_id,
        COALESCE(name, '') as item_title,
        COALESCE(last_update, '1970-01-01'::timestamp) as last_activity
    FROM contacts
    WHERE name ILIKE '%' || sqlc.arg(query)::text || '%'

    UNION ALL

    -- Rooms (conversations)
    SELECT
        'conversation'::text as item_type,
        room_id::text as item_id,
        COALESCE(user_defined_name, display_name, '') as item_title,
        COALESCE(last_activity, '1970-01-01'::timestamp) as last_activity
    FROM rooms
    WHERE COALESCE(user_defined_name, display_name) ILIKE '%' || sqlc.arg(query)::text || '%'

    UNION ALL

    -- Bookmarks
    SELECT
        'bookmark'::text as item_type,
        bt.bookmark_id::text as item_id,
        COALESCE(bt.title, '') as item_title,
        COALESCE(b.creation_date, '1970-01-01'::timestamp) as last_activity
    FROM bookmark_titles bt
    LEFT JOIN bookmarks b ON bt.bookmark_id = b.bookmark_id
    WHERE bt.title ILIKE '%' || sqlc.arg(query)::text || '%'

    UNION ALL

    -- Browser history
    SELECT
        'history'::text as item_type,
        id::text as item_id,
        COALESCE(title, '') as item_title,
        COALESCE(visit_date, '1970-01-01'::timestamp) as last_activity
    FROM browser_history
    WHERE title ILIKE '%' || sqlc.arg(query)::text || '%'

    UNION ALL

    -- Items (notes)
    SELECT
        'note'::text as item_type,
        id::text as item_id,
        COALESCE(title, '') as item_title,
        COALESCE(to_timestamp(modified)::timestamp, '1970-01-01'::timestamp) as last_activity
    FROM items
    WHERE title ILIKE '%' || sqlc.arg(query)::text || '%'
)
SELECT
    item_type,
    item_id,
    item_title,
    last_activity,
    (
        -- Exact match weight
        CASE
            WHEN lower(LEFT(item_title, 255)) = lower(LEFT(sqlc.arg(query)::text, 255))
            THEN sqlc.arg(exact_match_weight)::float
            ELSE 0
        END
    )
    +
    (
        -- Similarity weight using pg_trgm
        similarity(lower(LEFT(item_title, 255)), lower(LEFT(sqlc.arg(query)::text, 255))) * sqlc.arg(similarity_weight)::float
    )
    +
    (
        -- Recency weight
        CASE
            WHEN last_activity >= now() - interval '7 days'
                THEN sqlc.arg(recency_weight)::float
            WHEN last_activity >= now() - interval '30 days'
                THEN sqlc.arg(recency_weight)::float / 2.0
            ELSE 0
        END
    ) as search_score
FROM search_union
ORDER BY search_score DESC
LIMIT sqlc.arg(result_limit)::int;

-- name: GetSimilarQuestions :many
WITH similar_qa AS (
    SELECT
        bcr.content as question,
        bcr.bookmark_id,
        bcr.strategy,
        1 - (bcr.embedding <=> sqlc.arg(embedding)::vector) as similarity
    FROM bookmark_content_references as bcr
    WHERE bcr.strategy = 'qa-v2-passage'
    ORDER BY similarity DESC
    LIMIT sqlc.arg(search_limit)::int
)
SELECT
    b.bookmark_id,
    COALESCE(bt.title, 'Untitled Bookmark') as title,
    b.url,
    COALESCE(reader.processed_content, '') as summary,
    sqa.question,
    sqa.similarity,
    sqa.strategy
FROM similar_qa sqa
INNER JOIN bookmarks b ON sqa.bookmark_id = b.bookmark_id
LEFT JOIN bookmark_titles bt ON b.bookmark_id = bt.bookmark_id
LEFT JOIN processed_contents reader ON b.bookmark_id = reader.bookmark_id AND reader.strategy_used = 'reader'
ORDER BY sqa.similarity DESC;
