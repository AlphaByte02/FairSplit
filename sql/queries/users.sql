-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = $1;


-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = $1;


-- name: GetUserByUsername :one
SELECT
    *
FROM
    users
WHERE
    LOWER(username) = LOWER($1);


-- name: GetUserByEmailOrUsername :one
SELECT
    *
FROM
    users
WHERE
    email = $1
    OR LOWER(username) = LOWER($1);


-- name: UpdateUser :exec
UPDATE users
SET
    username = $2,
    paypal_username = $3,
    iban = $4
WHERE
    id = $1;


-- name: UpdateUserPicture :exec
UPDATE users
SET
    picture = $2
WHERE
    id = $1;


-- name: CreateUser :one
INSERT INTO
    users (id, email, username, picture)
VALUES
    ($1, $2, $3, $4)
RETURNING
    *;


-- name: DeleteUser :exec
DELETE FROM users
WHERE
    id = $1;


-- name: CheckUserExists :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            users
        WHERE
            LOWER(username) = LOWER($1)
    );
