-- name: GetMessage :one
SELECT
    m.message_id,
    m.sender_contact_id,
    m.room_id,
    m.event_id,
    m.event_datetime,
    m.body,
    m.formatted_body,
    m.message_type,
    c.name AS sender_name,
    c.email AS sender_email
FROM messages m
LEFT JOIN contacts c ON m.sender_contact_id = c.contact_id
WHERE m.message_id = $1;

-- name: GetRoomMessages :many
SELECT
    m.message_id,
    m.sender_contact_id,
    m.event_id,
    m.event_datetime,
    m.body,
    m.formatted_body,
    m.message_type,
    m.message_classification,
    o.data AS transcription_data,
    b.bookmark_id,
    b.url AS bookmark_url,
    bt.title AS bookmark_title,
    bcr.content AS bookmark_summary
FROM messages m
LEFT JOIN observations o ON o.ref = m.message_id AND o.type = 'transcription'
LEFT JOIN bookmark_sources bs ON bs.source_uri = m.event_id
LEFT JOIN bookmarks b ON b.bookmark_id = bs.bookmark_id
LEFT JOIN bookmark_titles bt ON bt.bookmark_id = b.bookmark_id
LEFT JOIN bookmark_content_references bcr ON bcr.bookmark_id = b.bookmark_id
    AND bcr.strategy = 'summary-reader'
WHERE m.room_id = $1
    AND CASE
        WHEN sqlc.narg('before_datetime')::timestamp IS NOT NULL THEN
            m.event_datetime < sqlc.narg('before_datetime')::timestamp
        ELSE TRUE
    END
ORDER BY m.event_datetime DESC
LIMIT $2 OFFSET $3;

-- name: SearchRoomMessages :many
SELECT
    m.message_id,
    m.sender_contact_id,
    m.event_id,
    m.event_datetime,
    m.body,
    m.formatted_body,
    m.message_type,
    m.message_classification,
    o.data AS transcription_data,
    b.bookmark_id,
    b.url AS bookmark_url,
    bt.title AS bookmark_title,
    bcr.content AS bookmark_summary
FROM messages m
LEFT JOIN observations o ON o.ref = m.message_id AND o.type = 'transcription'
LEFT JOIN bookmark_sources bs ON bs.source_uri = m.event_id
LEFT JOIN bookmarks b ON b.bookmark_id = bs.bookmark_id
LEFT JOIN bookmark_titles bt ON bt.bookmark_id = b.bookmark_id
LEFT JOIN bookmark_content_references bcr ON bcr.bookmark_id = b.bookmark_id
    AND bcr.strategy = 'summary-reader'
WHERE m.room_id = $1
    AND m.body ILIKE '%' || $2 || '%'
ORDER BY m.event_datetime DESC
LIMIT $3 OFFSET $4;

-- name: SearchMessages :many
SELECT
    m.message_id,
    m.sender_contact_id,
    m.room_id,
    m.event_id,
    m.event_datetime,
    m.body,
    m.formatted_body,
    m.message_type,
    mtr.text_content,
    ts_rank(mtr.search_vector, to_tsquery('english', $1)) AS rank
FROM messages m
INNER JOIN message_text_representation mtr ON m.message_id = mtr.message_id
WHERE mtr.search_vector @@ to_tsquery('english', $1)
ORDER BY rank DESC, m.event_datetime DESC
LIMIT $2 OFFSET $3;

-- name: GetMessageContacts :many
SELECT DISTINCT
    c.contact_id,
    c.name,
    c.email
FROM contacts c
WHERE c.contact_id = ANY($1::uuid[]);

-- name: GetBeforeMessageDatetime :one
SELECT event_datetime
FROM messages
WHERE message_id = $1;

-- name: GetContextMessages :many
WITH matched_messages AS (
    SELECT message_id, event_datetime, room_id
    FROM messages
    WHERE message_id = ANY($2::uuid[])
        AND messages.room_id = $1
),
context_messages AS (
    SELECT DISTINCT m.message_id
    FROM messages m
    CROSS JOIN matched_messages mm
    WHERE m.room_id = $1
        AND (
            (m.event_datetime < mm.event_datetime
             AND m.message_id IN (
                 SELECT message_id
                 FROM messages
                 WHERE room_id = $1
                     AND event_datetime < mm.event_datetime
                 ORDER BY event_datetime DESC
                 LIMIT $3
             ))
            OR
            (m.event_datetime > mm.event_datetime
             AND m.message_id IN (
                 SELECT message_id
                 FROM messages
                 WHERE room_id = $1
                     AND event_datetime > mm.event_datetime
                 ORDER BY event_datetime ASC
                 LIMIT $4
             ))
        )
)
SELECT
    m.message_id,
    m.sender_contact_id,
    m.event_id,
    m.event_datetime,
    m.body,
    m.formatted_body,
    m.message_type,
    m.message_classification,
    o.data AS transcription_data,
    b.bookmark_id,
    b.url AS bookmark_url,
    bt.title AS bookmark_title,
    bcr.content AS bookmark_summary
FROM messages m
LEFT JOIN observations o ON o.ref = m.message_id AND o.type = 'transcription'
LEFT JOIN bookmark_sources bs ON bs.source_uri = m.event_id
LEFT JOIN bookmarks b ON b.bookmark_id = bs.bookmark_id
LEFT JOIN bookmark_titles bt ON bt.bookmark_id = b.bookmark_id
LEFT JOIN bookmark_content_references bcr ON bcr.bookmark_id = b.bookmark_id
    AND bcr.strategy = 'summary-reader'
WHERE m.message_id IN (SELECT message_id FROM context_messages)
ORDER BY m.event_datetime DESC;

-- name: GetAllMessageContents :many
SELECT content
FROM raw_messages;

-- name: GetMessageTextRepresentations :many
SELECT
    id,
    message_id,
    text_content,
    source_type,
    created_at
FROM message_text_representation
WHERE message_id = $1;
