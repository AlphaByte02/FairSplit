-- name: GetSession :one
SELECT
    s.*,
    sqlc.embed(u)
FROM
    sessions s
    JOIN users u ON u.id = s.created_by_id
WHERE
    s.id = $1;

-- name: ListSessionsForUser :many
SELECT
    s.*
FROM
    sessions s
    JOIN session_participants sp ON s.id = sp.session_id
WHERE
    sp.user_id = $1;

-- name: CreateSession :one
INSERT INTO
    sessions (id, created_by_id, name)
VALUES
    ($1, $2, $3)
RETURNING
    *;

-- name: AddParticipant :exec
INSERT INTO
    session_participants (session_id, user_id)
VALUES
    ($1, $2);

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE
    id = $1;

-- name: CloseSession :exec
UPDATE sessions
SET
    is_closed = true
WHERE
    id = $1;
