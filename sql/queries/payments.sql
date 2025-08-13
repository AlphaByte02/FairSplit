-- name: ListPaymentsForSession :many
SELECT
    *
FROM
    payments
WHERE
    session_id = $1
ORDER BY
    created_at;


-- name: CreatePayment :one
INSERT INTO
    payments (id, session_id, payer_id, amount, description)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;


-- name: AddPaymentParticipant :exec
INSERT INTO
    payment_participants (payment_id, user_id)
VALUES
    ($1, $2);
