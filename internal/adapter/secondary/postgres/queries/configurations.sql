-- name: ListConfigurations :many
SELECT config_id, key, value, is_secret, created_at, updated_at
FROM configurations
WHERE
    (sqlc.narg('key_prefix')::text IS NULL OR key LIKE sqlc.narg('key_prefix')::text || '%')
    AND (sqlc.narg('include_secrets')::boolean = true OR is_secret = false)
ORDER BY key;

-- name: GetConfigurationByKey :one
SELECT config_id, key, value, is_secret, created_at, updated_at
FROM configurations
WHERE key = $1;

-- name: GetConfigurationsByPrefix :many
SELECT config_id, key, value, is_secret, created_at, updated_at
FROM configurations
WHERE
    key LIKE $1 || '%'
    AND (sqlc.narg('include_secrets')::boolean = true OR is_secret = false)
ORDER BY key;

-- name: CreateConfiguration :one
INSERT INTO configurations (key, value, is_secret, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING config_id, key, value, is_secret, created_at, updated_at;

-- name: UpdateConfiguration :one
UPDATE configurations
SET value = $2, is_secret = $3, updated_at = $4
WHERE key = $1
RETURNING config_id, key, value, is_secret, created_at, updated_at;

-- name: DeleteConfiguration :exec
DELETE FROM configurations
WHERE key = $1;

-- name: UpsertConfiguration :one
INSERT INTO configurations (key, value, is_secret, updated_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, is_secret = EXCLUDED.is_secret, updated_at = EXCLUDED.updated_at
RETURNING config_id, key, value, is_secret, created_at, updated_at;
