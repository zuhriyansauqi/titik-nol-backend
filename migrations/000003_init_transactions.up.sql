-- 000003_init_transactions.up.sql
CREATE TYPE tx_type_enum AS ENUM ('INCOME', 'EXPENSE', 'TRANSFER', 'ADJUSTMENT');

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL, -- Nullable untuk Quick-Log
    transaction_type tx_type_enum NOT NULL,
    amount BIGINT NOT NULL,
    note TEXT,
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indeks untuk optimasi query Dashboard (PRD Metrik)
CREATE INDEX idx_transactions_user_date ON transactions(user_id, transaction_date);
CREATE INDEX idx_accounts_user ON accounts(user_id);
