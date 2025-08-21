-- name: GetTransaction :one
SELECT
    *
FROM
    transactions
WHERE
    id = $1;


-- name: ListTransactionsBySession :many
SELECT
    t.*,
    /* sql-formatter-disable */
    sqlc.embed(payer),
    /* sql-formatter-enable */
    ARRAY_AGG(
        u.username
        ORDER BY
            LOWER(u.username)
    )::TEXT[] AS participants
FROM
    transactions t
    JOIN users payer ON payer.id = t.payer_id
    LEFT JOIN transaction_participants tp ON tp.transaction_id = t.id
    LEFT JOIN users u ON u.id = tp.user_id
WHERE
    t.session_id = $1
GROUP BY
    t.id,
    t.session_id,
    t.payer_id,
    t.amount,
    t.description,
    t.created_at,
    t.updated_at,
    payer.id,
    payer.username,
    payer.created_at,
    payer.updated_at
ORDER BY
    t.created_at DESC;


-- name: ListTransactions :many
SELECT
    /* sql-formatter-disable */
    sqlc.embed(t),
    sqlc.embed(s)
    /* sql-formatter-enable */
FROM
    transactions t
    JOIN sessions s ON s.id = t.session_id
ORDER BY
    t.created_at DESC;


-- name: UpdateTransactions :exec
UPDATE transactions
SET
    session_id = $2,
    payer_id = $3,
    amount = $4,
    description = $5
WHERE
    id = $1;


-- name: CreateTransaction :one
INSERT INTO
    transactions (id, session_id, payer_id, amount, description, created_by_id)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING
    *;


-- name: AddTransactionParticipant :exec
INSERT INTO
    transaction_participants (transaction_id, user_id)
VALUES
    ($1, $2);


-- name: ListTransactionParticipants :many
SELECT
    u.*
FROM
    transaction_participants tp
    JOIN users u ON u.id = tp.user_id
WHERE
    transaction_id = $1
ORDER BY
    LOWER(u.username);


-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE
    id = $1;


-- name: CountTransactionByUserAndSession :one
SELECT
    COUNT(tp.*) AS "count"
FROM
    transaction_participants tp
    JOIN transactions t ON t.id = tp.transaction_id
WHERE
    tp.user_id = $1
    AND t.session_id = $2;
