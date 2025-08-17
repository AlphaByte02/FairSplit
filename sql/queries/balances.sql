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


-- TODO
-- name: GetFinalBalances :many
SELECT
    ts.id_soggetto AS id_debitore,
    t.id_pagante AS id_creditore,
    SUM(t.importo::numeric / cnt.tot_partecipanti) AS spesa
FROM
    transazioni t
    JOIN soggetti ts ON ts.id_transazione = t.id
    AND ts.id_soggetto <> t.id_pagante
    JOIN (
        SELECT
            s2.id_transazione,
            COUNT(*) AS tot_partecipanti
        FROM
            soggetti s2
        GROUP BY
            s2.id_transazione
    ) cnt ON cnt.id_transazione = t.id
GROUP BY
    ts.id_soggetto,
    t.id_pagante
