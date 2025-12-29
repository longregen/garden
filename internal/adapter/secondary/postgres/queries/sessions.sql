-- name: SearchSessionsWithEmbeddings :many
SELECT
    s.session_id,
    s.room_id,
    s.first_date_time,
    s.last_date_time,
    ss.summary,
    r.display_name,
    r.user_defined_name,
    1 - (ss.embedding <=> $1::vector) as similarity
FROM session_summaries ss
JOIN sessions s ON ss.session_id = s.session_id
JOIN rooms r ON s.room_id = r.room_id
WHERE ss.summary <> '' AND ss.summary IS NOT NULL
ORDER BY similarity DESC
LIMIT $2;

-- name: SearchSessionsWithText :many
SELECT
    s.session_id,
    s.room_id,
    s.first_date_time,
    s.last_date_time,
    ss.summary,
    r.display_name,
    r.user_defined_name
FROM sessions s
LEFT JOIN session_summaries ss ON s.session_id = ss.session_id
LEFT JOIN rooms r ON s.room_id = r.room_id
WHERE ss.summary ILIKE $1
ORDER BY s.last_date_time DESC
LIMIT $2;

-- name: GetSessionSummaries :many
SELECT
    s.session_id,
    s.first_date_time,
    s.last_date_time,
    ss.summary
FROM sessions s
LEFT JOIN session_summaries ss ON s.session_id = ss.session_id
WHERE s.room_id = $1
ORDER BY s.first_date_time DESC;

-- name: SearchContactSessionSummaries :many
SELECT
    s.session_id,
    ss.summary,
    s.first_date_time,
    s.last_date_time,
    s.room_id
FROM session_summaries ss
LEFT JOIN sessions s ON ss.session_id = s.session_id
LEFT JOIN room_participants rp ON rp.room_id = s.room_id
WHERE rp.contact_id = $1
  AND ss.summary ILIKE $2
ORDER BY s.last_date_time DESC;

-- name: GetSessionsForRoom :many
SELECT
    s.session_id,
    s.first_date_time,
    s.last_date_time,
    COUNT(sm.message_id) as message_count
FROM sessions s
LEFT JOIN session_message sm ON s.session_id = sm.session_id
WHERE s.room_id = $1
GROUP BY s.session_id, s.first_date_time, s.last_date_time
ORDER BY s.first_date_time ASC;

-- name: GetSessionParticipantActivity :many
SELECT
    m.sender_contact_id,
    ct.name as sender_name,
    s.session_id,
    s.first_date_time,
    COUNT(m.message_id) as message_count
FROM messages m
INNER JOIN session_message sm ON m.message_id = sm.message_id
INNER JOIN sessions s ON sm.session_id = s.session_id
INNER JOIN contacts ct ON m.sender_contact_id = ct.contact_id
WHERE s.room_id = $1
GROUP BY m.sender_contact_id, ct.name, s.session_id, s.first_date_time
ORDER BY s.first_date_time ASC;

-- name: GetSessionMessages :many
SELECT
    m.message_id,
    m.sender_contact_id,
    m.body,
    m.formatted_body,
    m.message_type,
    m.message_classification,
    m.event_datetime,
    o.data as transcription_data
FROM session_message sm
INNER JOIN messages m ON sm.message_id = m.message_id
LEFT JOIN observations o ON o.ref = m.message_id AND o.type = 'transcription'
WHERE sm.session_id = $1
ORDER BY m.event_datetime ASC;

-- name: GetContactsByIds :many
SELECT
    contact_id,
    name
FROM contacts
WHERE contact_id = ANY($1::uuid[]);
