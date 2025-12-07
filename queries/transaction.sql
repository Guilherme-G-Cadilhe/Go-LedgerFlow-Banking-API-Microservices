-- name: CreateTransaction :one
INSERT INTO transactions (
    from_wallet_id,
    to_wallet_id,
    amount,
    status,
    idempotency_key
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE from_wallet_id = $1 OR to_wallet_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;