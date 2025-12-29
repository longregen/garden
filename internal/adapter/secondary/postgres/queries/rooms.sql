-- name: ListRooms :many
SELECT
    r.room_id,
    r.display_name,
    r.user_defined_name,
    r.last_activity,
    COUNT(DISTINCT rp.contact_id)::int AS participant_count,
    MAX(m.created_at) AS last_message_time
FROM rooms r
LEFT JOIN room_participants rp ON r.room_id = rp.room_id
LEFT JOIN messages m ON r.room_id = m.room_id
WHERE
    CASE
        WHEN sqlc.narg('search_text')::text IS NOT NULL THEN
            (r.display_name ILIKE '%' || sqlc.narg('search_text')::text || '%' OR
             r.user_defined_name ILIKE '%' || sqlc.narg('search_text')::text || '%')
        ELSE TRUE
    END
GROUP BY r.room_id
ORDER BY MAX(m.created_at) DESC NULLS LAST
LIMIT $1 OFFSET $2;

-- name: CountRooms :one
SELECT COUNT(DISTINCT r.room_id)
FROM rooms r
WHERE
    CASE
        WHEN sqlc.narg('search_text')::text IS NOT NULL THEN
            (r.display_name ILIKE '%' || sqlc.narg('search_text')::text || '%' OR
             r.user_defined_name ILIKE '%' || sqlc.narg('search_text')::text || '%')
        ELSE TRUE
    END;

-- name: GetRoom :one
SELECT
    r.room_id,
    r.display_name,
    r.user_defined_name,
    r.source_id,
    r.last_activity,
    MAX(m.created_at) AS last_message_time
FROM rooms r
LEFT JOIN messages m ON r.room_id = m.room_id
WHERE r.room_id = $1
GROUP BY r.room_id;

-- name: GetRoomParticipants :many
SELECT
    rp.room_id,
    rp.contact_id,
    rp.known_last_presence,
    rp.known_last_exit,
    c.name AS contact_name,
    c.email AS contact_email
FROM room_participants rp
INNER JOIN contacts c ON rp.contact_id = c.contact_id
WHERE rp.room_id = $1;

-- name: UpdateRoomName :exec
UPDATE rooms
SET user_defined_name = $2
WHERE room_id = $1;

-- name: GetRoomSessionsCount :one
SELECT COUNT(session_id)::int
FROM sessions
WHERE room_id = $1;
