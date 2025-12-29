-- name: GetContactStats :one
SELECT
    COUNT(*)::bigint as total,
    COUNT(*) FILTER (WHERE last_update >= now() - interval '30 days')::bigint as recently_active,
    COUNT(*) FILTER (WHERE last_update >= now() - interval '30 days')::bigint as current_period_count,
    COUNT(*) FILTER (WHERE last_update >= now() - interval '60 days' AND last_update < now() - interval '30 days')::bigint as previous_period_count
FROM contacts;

-- name: GetSessionStats :one
SELECT
    COUNT(*)::bigint as total,
    COUNT(*) FILTER (WHERE last_date_time >= now() - interval '30 days')::bigint as recent_count,
    COUNT(*) FILTER (WHERE last_date_time >= now() - interval '30 days')::bigint as current_period_count,
    COUNT(*) FILTER (WHERE last_date_time >= now() - interval '60 days' AND last_date_time < now() - interval '30 days')::bigint as previous_period_count
FROM sessions;

-- name: GetBookmarkStats :one
SELECT
    COUNT(*)::bigint as total,
    COUNT(*) FILTER (WHERE creation_date >= now() - interval '30 days')::bigint as recent_count,
    COUNT(*) FILTER (WHERE creation_date >= now() - interval '30 days')::bigint as current_period_count,
    COUNT(*) FILTER (WHERE creation_date >= now() - interval '60 days' AND creation_date < now() - interval '30 days')::bigint as previous_period_count
FROM bookmarks;

-- name: GetHistoryStats :one
SELECT
    COUNT(*)::bigint as total,
    COUNT(*) FILTER (WHERE visit_date >= now() - interval '30 days')::bigint as recent_count,
    COUNT(*) FILTER (WHERE visit_date >= now() - interval '30 days')::bigint as current_period_count,
    COUNT(*) FILTER (WHERE visit_date >= now() - interval '60 days' AND visit_date < now() - interval '30 days')::bigint as previous_period_count
FROM browser_history;

-- name: GetRecentContacts :many
SELECT
    contact_id::text as id,
    'Contact'::text as category,
    name,
    last_update as date
FROM contacts
WHERE last_update >= now() - interval '30 days'
ORDER BY last_update DESC
LIMIT 5;

-- name: GetRecentSessions :many
SELECT
    s.session_id::text as id,
    'Session'::text as category,
    COALESCE(r.display_name, '') as name,
    s.last_date_time as date
FROM sessions s
LEFT JOIN rooms r ON r.room_id = s.room_id
WHERE s.last_date_time >= now() - interval '30 days'
ORDER BY s.last_date_time DESC
LIMIT 5;

-- name: GetRecentBookmarks :many
SELECT
    b.bookmark_id::text as id,
    'Bookmark'::text as category,
    COALESCE(bt.title, '') as name,
    b.creation_date as date
FROM bookmarks b
LEFT JOIN bookmark_titles bt ON bt.bookmark_id = b.bookmark_id
WHERE b.creation_date >= now() - interval '30 days'
ORDER BY b.creation_date DESC
LIMIT 5;

-- name: GetRecentHistory :many
SELECT
    browser_history.id::text as id,
    'History'::text as category,
    COALESCE(browser_history.title, '') as name,
    browser_history.visit_date as date
FROM browser_history
WHERE browser_history.visit_date >= now() - interval '30 days'
ORDER BY browser_history.visit_date DESC
LIMIT 5;

-- name: GetRecentNotes :many
SELECT
    id::text as id,
    'Note'::text as category,
    COALESCE(title, '') as name,
    to_timestamp(modified) as date
FROM items
WHERE to_timestamp(modified) >= now() - interval '30 days'
ORDER BY to_timestamp(modified) DESC
LIMIT 5;
