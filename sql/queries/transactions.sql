-- name: ListTransactionsBySession :many
SELECT
    t.*,
    /* sql-formatter-disable */
    sqlc.embed(payer),
    /* sql-formatter-enable */
    ARRAY_AGG(DISTINCT u.username)::TEXT[] AS participants
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


-- name: CreateTransaction :one
INSERT INTO
    transactions (id, session_id, payer_id, amount, description)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;


-- name: AddTransactionParticipant :exec
INSERT INTO
    transaction_participants (transaction_id, user_id)
VALUES
    ($1, $2);


-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE
    id = $1;


-- name: CountTransactionByUser :one
SELECT
    COUNT(tp.*) AS "count"
FROM
    transaction_participants tp
WHERE
    tp.user_id = $1;
