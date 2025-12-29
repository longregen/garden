-- name: ListSocialPosts :many
SELECT
    post_id,
    content,
    twitter_post_id,
    bluesky_post_id,
    created_at,
    updated_at,
    status,
    error_message
FROM social_posts
WHERE ($1::text IS NULL OR status = $1)
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: CountSocialPosts :one
SELECT COUNT(*)
FROM social_posts
WHERE ($1::text IS NULL OR status = $1);

-- name: GetSocialPost :one
SELECT
    post_id,
    content,
    twitter_post_id,
    bluesky_post_id,
    created_at,
    updated_at,
    status,
    error_message
FROM social_posts
WHERE post_id = $1;

-- name: CreateSocialPost :one
INSERT INTO social_posts (content, status)
VALUES ($1, $2)
RETURNING post_id, content, twitter_post_id, bluesky_post_id, created_at, updated_at, status, error_message;

-- name: UpdateSocialPostStatus :exec
UPDATE social_posts
SET
    status = $1,
    error_message = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE post_id = $3;

-- name: UpdateSocialPostTwitterID :exec
UPDATE social_posts
SET
    twitter_post_id = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE post_id = $2;

-- name: UpdateSocialPostBlueskyID :exec
UPDATE social_posts
SET
    bluesky_post_id = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE post_id = $2;

-- name: DeleteSocialPost :exec
DELETE FROM social_posts
WHERE post_id = $1;
