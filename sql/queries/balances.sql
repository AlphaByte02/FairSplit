-- name: GetIntermediateBalances :many
SELECT
    /* sql-formatter-disable */
    sqlc.embed(t),
    sqlc.embed(payer),
    sqlc.embed(dep),
    /* sql-formatter-enable */
    (
        t.amount / COUNT(tp.*) OVER (
            PARTITION BY
                t.id
        )
    )::numeric(12, 2) AS amount_per_user
FROM
    transactions t
    LEFT JOIN users payer ON t.payer_id = payer.id
    LEFT JOIN transaction_participants tp ON tp.transaction_id = t.id
    LEFT JOIN users dep ON dep.id = tp.user_id
WHERE
    t.session_id = $1
ORDER BY
    t.id;


-- name: GetSessionBalances :many
WITH
    tx_info AS (
        SELECT
            id,
            amount,
            (
                SELECT
                    COUNT(*)
                FROM
                    transaction_participants tp
                WHERE
                    tp.transaction_id = t.id
            ) AS num_parts
        FROM
            transactions t
        WHERE
            t.session_id = $1
    ),
    paid AS (
        SELECT
            t.payer_id AS user_id,
            SUM(t.amount) AS total_paid
        FROM
            transactions t
        WHERE
            t.session_id = $1
        GROUP BY
            t.payer_id
    ),
    consumed AS (
        SELECT
            tp.user_id,
            SUM(tx.amount / tx.num_parts::numeric) AS total_consumed
        FROM
            transaction_participants tp
            JOIN tx_info tx ON tx.id = tp.transaction_id
        GROUP BY
            tp.user_id
    ),
    balances AS (
        SELECT
            sp.user_id,
            CAST(COALESCE(p.total_paid, 0) - COALESCE(c.total_consumed, 0) AS numeric(12, 2)) AS balance
        FROM
            session_participants sp
            LEFT JOIN paid p ON p.user_id = sp.user_id
            LEFT JOIN consumed c ON c.user_id = sp.user_id
        WHERE
            sp.session_id = $1
    )
SELECT
    /* sql-formatter-disable */
    sqlc.embed(u),
    /* sql-formatter-enable */
    b.balance
FROM
    balances b
    LEFT JOIN users u ON b.user_id = u.id;


-- name: SaveFinalBalance :batchone
INSERT INTO
    final_balances (id, session_id, creditor_id, debtor_id, amount)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;


-- name: GetFinalBalancesBySession :many
SELECT
    fb.*,
    /* sql-formatter-disable */
    sqlc.embed(cred),
    sqlc.embed(debt)
    /* sql-formatter-enable */
FROM
    final_balances fb
    LEFT JOIN users cred ON fb.creditor_id = cred.id
    LEFT JOIN users debt ON fb.debtor_id = debt.id
WHERE
    session_id = $1;


-- name: ToggleDeptPaid :exec
UPDATE final_balances
SET
    is_paid = NOT is_paid
WHERE
    id = $1;
