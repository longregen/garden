-- name: ListBrowserHistory :many
SELECT
    browser_history.id,
    browser_history.url,
    browser_history.title,
    browser_history.visit_date,
    browser_history.typed,
    browser_history.hidden,
    browser_history.imported_from_firefox_place_id,
    browser_history.imported_from_firefox_visit_id,
    browser_history.domain,
    browser_history.created_at
FROM browser_history
WHERE
    (sqlc.narg(search_query)::text IS NULL OR (
        browser_history.title ILIKE '%' || sqlc.narg(search_query)::text || '%' OR
        browser_history.url ILIKE '%' || sqlc.narg(search_query)::text || '%'
    ))
    AND (sqlc.narg(start_date)::timestamp IS NULL OR browser_history.visit_date >= sqlc.narg(start_date)::timestamp)
    AND (sqlc.narg(end_date)::timestamp IS NULL OR browser_history.visit_date <= sqlc.narg(end_date)::timestamp)
    AND (sqlc.narg(domain)::text IS NULL OR browser_history.domain = sqlc.narg(domain)::text)
ORDER BY browser_history.visit_date DESC
LIMIT sqlc.arg(page_size)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: CountBrowserHistory :one
SELECT COUNT(*)::bigint
FROM browser_history
WHERE
    (sqlc.narg(search_query)::text IS NULL OR (
        browser_history.title ILIKE '%' || sqlc.narg(search_query)::text || '%' OR
        browser_history.url ILIKE '%' || sqlc.narg(search_query)::text || '%'
    ))
    AND (sqlc.narg(start_date)::timestamp IS NULL OR browser_history.visit_date >= sqlc.narg(start_date)::timestamp)
    AND (sqlc.narg(end_date)::timestamp IS NULL OR browser_history.visit_date <= sqlc.narg(end_date)::timestamp)
    AND (sqlc.narg(domain)::text IS NULL OR browser_history.domain = sqlc.narg(domain)::text);

-- name: GetTopDomains :many
SELECT
    domain,
    COUNT(*)::bigint as visit_count
FROM browser_history
WHERE domain IS NOT NULL
GROUP BY domain
ORDER BY visit_count DESC
LIMIT sqlc.arg(domain_limit)::int;

-- name: GetRecentBrowserHistory :many
SELECT
    browser_history.id,
    browser_history.url,
    browser_history.title,
    browser_history.visit_date,
    browser_history.typed,
    browser_history.hidden,
    browser_history.imported_from_firefox_place_id,
    browser_history.imported_from_firefox_visit_id,
    browser_history.domain,
    browser_history.created_at
FROM browser_history
ORDER BY browser_history.visit_date DESC
LIMIT sqlc.arg(history_limit)::int;
