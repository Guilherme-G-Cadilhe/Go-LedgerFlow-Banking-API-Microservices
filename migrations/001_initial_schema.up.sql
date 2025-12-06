-- migrations/001_initial_schema.sql

-- 1. Wallets
CREATE TABLE IF NOT EXISTS wallets (
    id BIGSERIAL PRIMARY KEY,
    -- Usamos BIGINT para armazenar centavos Ex: R$ 100.00 é armazenado como 10000
    -- CHECK constraint garante que o saldo nunca seja negativo no nível do banco.
    balance BIGINT NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
    
    -- Version para Optimistic Locking (segurança extra)
    version INT NOT NULL DEFAULT 1,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index para lock rápido
CREATE INDEX idx_wallets_version ON wallets(id, version);

-- 2. Transactions (Ledger imutável)
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_wallet_id BIGINT NOT NULL REFERENCES wallets(id),
    to_wallet_id BIGINT NOT NULL REFERENCES wallets(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    
    -- Idempotency Key única para evitar duplicação
    idempotency_key VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Regra de negócio: não pode transferir para si mesmo
    CONSTRAINT different_wallets CHECK (from_wallet_id != to_wallet_id)
);

CREATE INDEX idx_transactions_from ON transactions(from_wallet_id);
CREATE INDEX idx_transactions_to ON transactions(to_wallet_id);
-- Garante que a mesma chave de idempotência não seja usada 2x (UNIQUE)
CREATE UNIQUE INDEX idx_transactions_idem_key ON transactions(idempotency_key) WHERE idempotency_key IS NOT NULL;