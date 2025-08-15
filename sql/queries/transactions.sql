-- name: ListTransactionsBySession :many
SELECT
    t.*,
    /* sql-formatter-disable */
    sqlc.embed(payer)
    /* sql-formatter-enable */
FROM
    transactions t
    JOIN users AS payer ON payer.id = t.payer_id
WHERE
    t.session_id = $1
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
