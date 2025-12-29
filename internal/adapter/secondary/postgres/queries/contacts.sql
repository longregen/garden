-- name: GetContact :one
SELECT
    c.contact_id,
    c.name,
    c.email,
    c.phone,
    c.birthday,
    c.notes,
    c.extras,
    c.creation_date,
    c.last_update,
    cs.last_week_messages,
    cs.groups_in_common
FROM contacts c
LEFT JOIN contact_stats cs ON c.contact_id = cs.contact_id
WHERE c.contact_id = $1;

-- name: GetContactEvaluation :one
SELECT
    id,
    contact_id,
    importance,
    closeness,
    fondness,
    created_at,
    updated_at
FROM contact_evals
WHERE contact_id = $1;

-- name: GetContactTags :many
SELECT
    ctn.tag_id,
    ctn.name
FROM contact_tags ct
INNER JOIN contact_tagnames ctn ON ct.tag_id = ctn.tag_id
WHERE ct.contact_id = $1;

-- name: GetContactKnownNames :many
SELECT
    id,
    contact_id,
    name
FROM contact_known_names
WHERE contact_id = $1;

-- name: GetContactRooms :many
SELECT
    r.room_id,
    r.display_name,
    r.user_defined_name
FROM room_participants rp
INNER JOIN rooms r ON rp.room_id = r.room_id
WHERE rp.contact_id = $1;

-- name: GetContactSources :many
SELECT
    id,
    contact_id,
    source_id,
    source_name
FROM contact_sources
WHERE contact_id = $1;

-- name: ListContacts :many
SELECT
    c.contact_id,
    c.name,
    c.email,
    c.phone,
    c.birthday,
    c.notes,
    c.extras,
    c.creation_date,
    c.last_update,
    cs.last_week_messages,
    cs.groups_in_common
FROM contacts c
LEFT JOIN contact_stats cs ON c.contact_id = cs.contact_id
WHERE c.contact_id != '5f78a1e9-f68b-445a-8d97-63111a877fe0'
ORDER BY cs.last_week_messages DESC NULLS LAST
LIMIT $1 OFFSET $2;

-- name: SearchContacts :many
SELECT
    c.contact_id,
    c.name,
    c.email,
    c.phone,
    c.birthday,
    c.notes,
    c.extras,
    c.creation_date,
    c.last_update,
    cs.last_week_messages,
    cs.groups_in_common
FROM contacts c
LEFT JOIN contact_stats cs ON c.contact_id = cs.contact_id
WHERE c.contact_id != '5f78a1e9-f68b-445a-8d97-63111a877fe0'
  AND (c.name ILIKE $1 OR c.email ILIKE $1)
ORDER BY cs.last_week_messages DESC NULLS LAST
LIMIT $2 OFFSET $3;

-- name: CountContacts :one
SELECT COUNT(*) FROM contacts;

-- name: GetContactEvaluationsByContactIds :many
SELECT
    contact_id,
    importance,
    closeness,
    fondness
FROM contact_evals
WHERE contact_id = ANY($1::uuid[]);

-- name: GetContactTagsByContactIds :many
SELECT
    ct.contact_id,
    ctn.tag_id,
    ctn.name
FROM contact_tags ct
INNER JOIN contact_tagnames ctn ON ct.tag_id = ctn.tag_id
WHERE ct.contact_id = ANY($1::uuid[]);

-- name: CreateContact :one
INSERT INTO contacts (
    name,
    email,
    phone,
    birthday,
    notes,
    extras,
    creation_date,
    last_update
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
) RETURNING contact_id, name, email, phone, birthday, notes, extras, creation_date, last_update;

-- name: UpdateContact :exec
UPDATE contacts
SET
    name = COALESCE($2, name),
    email = COALESCE($3, email),
    phone = COALESCE($4, phone),
    birthday = COALESCE($5, birthday),
    notes = COALESCE($6, notes),
    extras = COALESCE($7, extras),
    last_update = NOW()
WHERE contact_id = $1;

-- name: UpdateContactFields :exec
UPDATE contacts
SET last_update = NOW()
WHERE contact_id = $1;

-- name: DeleteContact :exec
DELETE FROM contacts WHERE contact_id = $1;

-- name: CreateContactEvaluation :exec
INSERT INTO contact_evals (
    contact_id,
    importance,
    closeness,
    fondness,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
);

-- name: UpdateContactEvaluation :exec
UPDATE contact_evals
SET
    importance = COALESCE($2, importance),
    closeness = COALESCE($3, closeness),
    fondness = COALESCE($4, fondness),
    updated_at = NOW()
WHERE contact_id = $1;

-- name: GetContactEvaluationExists :one
SELECT EXISTS(SELECT 1 FROM contact_evals WHERE contact_id = $1);

-- name: MergeContacts :exec
CALL merge_contacts($1::uuid, $2::uuid);

-- name: GetAllTagNames :many
SELECT tag_id, name
FROM contact_tagnames
ORDER BY name;

-- name: GetContactTagByName :one
SELECT tag_id, name
FROM contact_tagnames
WHERE name = $1;

-- name: CreateTagName :one
INSERT INTO contact_tagnames (name, created_at)
VALUES ($1, NOW())
RETURNING tag_id, name;

-- name: GetContactTagExists :one
SELECT EXISTS(
    SELECT 1 FROM contact_tags
    WHERE contact_id = $1 AND tag_id = $2
);

-- name: AddContactTag :exec
INSERT INTO contact_tags (contact_id, tag_id)
VALUES ($1, $2);

-- name: RemoveContactTag :exec
DELETE FROM contact_tags
WHERE contact_id = $1 AND tag_id = $2;

-- name: EnsureContactStatsExist :exec
INSERT INTO contact_stats (contact_id)
SELECT contact_id FROM contacts
WHERE NOT EXISTS (
    SELECT 1 FROM contact_stats
    WHERE contact_stats.contact_id = contacts.contact_id
);

-- name: UpdateContactMessageStats :exec
UPDATE contact_stats
SET last_week_messages = (
    SELECT COUNT(message_id)
    FROM messages
    WHERE sender_contact_id = contact_stats.contact_id
      AND event_datetime > NOW() - INTERVAL '7 days'
);

-- name: ListAllContactSources :many
SELECT
    id,
    contact_id,
    source_id,
    source_name
FROM contact_sources
ORDER BY source_name;

-- name: CreateContactSource :one
INSERT INTO contact_sources (
    contact_id,
    source_id,
    source_name
) VALUES (
    $1, $2, $3
) RETURNING id, contact_id, source_id, source_name;

-- name: UpdateContactSource :one
UPDATE contact_sources
SET
    source_id = $2,
    source_name = $3
WHERE id = $1
RETURNING id, contact_id, source_id, source_name;

-- name: DeleteContactSource :exec
DELETE FROM contact_sources WHERE id = $1;
