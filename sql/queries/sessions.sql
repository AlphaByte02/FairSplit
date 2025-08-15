-- name: GetSession :one
SELECT
    /* sql-formatter-disable */
    sqlc.embed(s),
    sqlc.embed(u)
    /* sql-formatter-enable */
FROM
    sessions s
    JOIN users u ON u.id = s.created_by_id
WHERE
    s.id = $1;


-- name: ListSessionsForUser :many
SELECT
    *
FROM
    sessions
WHERE
    sessions.created_by_id = $1;


-- name: CreateSession :one
INSERT INTO
    sessions (id, created_by_id, name)
VALUES
    ($1, $2, $3)
RETURNING
    *;


-- name: ListSessionParticipants :many
SELECT
    u.*
FROM
    session_participants sp
    JOIN users u ON u.id = sp.user_id
WHERE
    session_id = $1
ORDER BY
    u.username;


-- name: AddSessionParticipant :exec
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
    is_closed = TRUE
WHERE
    id = $1;


-- name: CheckSessionAccess :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            session_participants
        WHERE
            session_id = $1
            AND user_id = $2
    ) AS has_access;
