-- name: CreateWallet :one
INSERT INTO wallets (balance)
VALUES ($1)
RETURNING *;

-- name: GetWallet :one
SELECT * FROM wallets
WHERE id = $1;

-- name: GetWalletForUpdate :one
-- 🚨 CRÍTICO: "FOR UPDATE" trava a linha até o fim da transação
SELECT * FROM wallets
WHERE id = $1
FOR UPDATE;

-- name: UpdateWalletBalance :exec
UPDATE wallets
SET balance = $2,
    version = version + 1,
    updated_at = NOW()
WHERE id = $1;

-- name: DebitWallet :execrows
-- Retorna número de linhas afetadas. Se 0, ou saldo insuficiente ou ID errado.
UPDATE wallets
SET balance = balance - sqlc.arg(amount),
    version = version + 1,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND balance >= sqlc.arg(amount); -- Segurança extra além do Check Constraint

-- name: CreditWallet :exec
UPDATE wallets
SET balance = balance + sqlc.arg(amount),
    version = version + 1,
    updated_at = NOW()
WHERE id = sqlc.arg(id);