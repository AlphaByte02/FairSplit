-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = $1;

-- name: GetUserByUsername :one
SELECT
    *
FROM
    users
WHERE
    username = $1;

-- name: CreateUser :one
INSERT INTO
    users (id, username)
VALUES
    ($1, $2)
RETURNING
    *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE
  id = $1;
