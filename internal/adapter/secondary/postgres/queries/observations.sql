-- name: CreateObservationWithReturn :one
INSERT INTO observations (data, type, source, tags, ref)
VALUES ($1, $2, $3, $4, $5)
RETURNING observation_id, data, type, source, tags, parent, ref, creation_date;

-- name: GetFeedbackStats :one
SELECT
    COUNT(CASE WHEN data->>'feedbackType' = 'upvote' THEN 1 END) as upvotes,
    COUNT(CASE WHEN data->>'feedbackType' = 'downvote' THEN 1 END) as downvotes,
    COUNT(CASE WHEN data->>'feedbackType' = 'trash' THEN 1 END) as trash
FROM observations
WHERE ref = $1 AND type = 'qa-feedback';

-- name: DeleteBookmarkContentReference :exec
DELETE FROM bookmark_content_references
WHERE id = $1 AND bookmark_id = $2;
