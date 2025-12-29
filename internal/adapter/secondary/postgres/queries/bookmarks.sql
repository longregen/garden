-- name: GetBookmark :one
SELECT
    bookmark_id,
    url,
    creation_date
FROM bookmarks
WHERE bookmark_id = $1;

-- name: ListBookmarks :many
SELECT DISTINCT
    b.bookmark_id,
    b.url,
    b.creation_date,
    bt.title
FROM bookmarks b
LEFT JOIN bookmark_titles bt ON b.bookmark_id = bt.bookmark_id
LEFT JOIN bookmark_category bc ON b.bookmark_id = bc.bookmark_id
WHERE
    ($1::uuid IS NULL OR bc.category_id = $1)
    AND ($2::text IS NULL OR bt.title ILIKE '%' || $2 || '%')
    AND ($3::timestamp IS NULL OR b.creation_date >= $3)
    AND ($4::timestamp IS NULL OR b.creation_date <= $4)
ORDER BY b.creation_date DESC
LIMIT $5
OFFSET $6;

-- name: CountBookmarks :one
SELECT COUNT(DISTINCT b.bookmark_id)
FROM bookmarks b
LEFT JOIN bookmark_titles bt ON b.bookmark_id = bt.bookmark_id
LEFT JOIN bookmark_category bc ON b.bookmark_id = bc.bookmark_id
WHERE
    ($1::uuid IS NULL OR bc.category_id = $1)
    AND ($2::text IS NULL OR bt.title ILIKE '%' || $2 || '%')
    AND ($3::timestamp IS NULL OR b.creation_date >= $3)
    AND ($4::timestamp IS NULL OR b.creation_date <= $4);

-- name: GetRandomBookmark :one
SELECT bookmark_id
FROM bookmarks
ORDER BY RANDOM()
LIMIT 1;

-- name: GetBookmarkDetails :one
SELECT
    b.bookmark_id,
    b.url,
    b.creation_date,
    c.name as category_name,
    bs.source_uri,
    bs.raw_source,
    bt.title,
    pc_lynx.processed_content as lynx_content,
    pc_reader.processed_content as reader_content,
    hr.status_code,
    hr.headers,
    hr.content as http_content,
    hr.fetch_date,
    bcr_summary.content as summary
FROM bookmarks b
LEFT JOIN bookmark_category bc ON b.bookmark_id = bc.bookmark_id
LEFT JOIN categories c ON bc.category_id = c.category_id
LEFT JOIN bookmark_sources bs ON b.bookmark_id = bs.bookmark_id
LEFT JOIN bookmark_titles bt ON b.bookmark_id = bt.bookmark_id
LEFT JOIN processed_contents pc_lynx ON b.bookmark_id = pc_lynx.bookmark_id AND pc_lynx.strategy_used = 'lynx'
LEFT JOIN processed_contents pc_reader ON b.bookmark_id = pc_reader.bookmark_id AND pc_reader.strategy_used = 'reader'
LEFT JOIN http_responses hr ON b.bookmark_id = hr.bookmark_id
LEFT JOIN bookmark_content_references bcr_summary ON b.bookmark_id = bcr_summary.bookmark_id AND bcr_summary.strategy = 'summary-reader'
WHERE b.bookmark_id = $1;

-- name: GetBookmarkQuestions :many
SELECT
    id,
    content
FROM bookmark_content_references
WHERE bookmark_id = $1 AND strategy = 'question-inference';

-- name: SearchSimilarBookmarks :many
SELECT
    b.bookmark_id,
    b.url,
    b.creation_date,
    bt.title,
    bcr_summary.content as summary
FROM bookmarks b
INNER JOIN bookmark_content_references bcr ON b.bookmark_id = bcr.bookmark_id
INNER JOIN bookmark_titles bt ON b.bookmark_id = bt.bookmark_id
LEFT JOIN bookmark_content_references bcr_summary ON b.bookmark_id = bcr_summary.bookmark_id AND bcr_summary.strategy = 'summary-reader'
WHERE bcr.strategy = $1
ORDER BY bcr.embedding <=> $2::vector
LIMIT $3;

-- name: UpdateBookmarkQuestion :exec
UPDATE bookmark_content_references
SET
    content = $1,
    embedding = $2::vector
WHERE id = $3 AND bookmark_id = $4;

-- name: DeleteBookmarkQuestion :exec
DELETE FROM bookmark_content_references
WHERE id = $1 AND bookmark_id = $2;

-- name: GetBookmarkForQuestion :one
SELECT
    b.bookmark_id,
    bt.title,
    bcr.content as summary
FROM bookmarks b
LEFT JOIN bookmark_titles bt ON b.bookmark_id = bt.bookmark_id
LEFT JOIN bookmark_content_references bcr ON b.bookmark_id = bcr.bookmark_id AND bcr.strategy = 'summary-reader'
WHERE b.bookmark_id = $1;

-- name: CreateObservation :exec
INSERT INTO observations (data, type, source, tags, ref)
VALUES ($1, $2, $3, $4, $5);

-- name: InsertHttpResponse :exec
INSERT INTO http_responses (bookmark_id, status_code, headers, content, fetch_date)
VALUES ($1, $2, $3, $4, $5);

-- name: GetLatestHttpResponse :one
SELECT
    response_id,
    status_code,
    headers,
    content,
    fetch_date
FROM http_responses
WHERE bookmark_id = $1
ORDER BY fetch_date DESC
LIMIT 1;

-- name: InsertProcessedContent :exec
INSERT INTO processed_contents (bookmark_id, strategy_used, processed_content)
VALUES ($1, $2, $3);

-- name: GetProcessedContentByStrategy :one
SELECT
    processed_content_id,
    processed_content
FROM processed_contents
WHERE bookmark_id = $1 AND strategy_used = $2
LIMIT 1;

-- name: CreateEmbeddingChunk :one
INSERT INTO bookmark_content_references (bookmark_id, content, strategy, embedding)
VALUES ($1, $2, $3, $4::vector)
RETURNING id;

-- name: GetBookmarkTitle :one
SELECT
    b.bookmark_id,
    b.url,
    b.creation_date,
    bt.title as existing_title,
    hr.content as raw_content,
    substring(pc.processed_content FROM '#([^\n]*)')::text as reader_title
FROM bookmarks b
LEFT JOIN bookmark_titles bt ON b.bookmark_id = bt.bookmark_id
LEFT JOIN http_responses hr ON b.bookmark_id = hr.bookmark_id
LEFT JOIN processed_contents pc ON b.bookmark_id = pc.bookmark_id AND pc.strategy_used = 'reader'
WHERE b.bookmark_id = $1;

-- name: InsertBookmarkTitle :exec
INSERT INTO bookmark_titles (bookmark_id, title, source)
VALUES ($1, $2, $3);

-- name: GetMissingHttpResponses :many
SELECT
    b.bookmark_id,
    b.url,
    b.creation_date
FROM bookmarks b
LEFT JOIN http_responses hr ON b.bookmark_id = hr.bookmark_id
WHERE hr.response_id IS NULL
ORDER BY b.creation_date DESC;

-- name: GetMissingReaderContent :many
SELECT
    b.bookmark_id,
    b.url,
    b.creation_date
FROM bookmarks b
LEFT JOIN processed_contents pc ON b.bookmark_id = pc.bookmark_id AND pc.strategy_used = 'reader'
WHERE pc.processed_content_id IS NULL
ORDER BY b.creation_date DESC;
