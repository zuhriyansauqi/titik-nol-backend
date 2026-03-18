-- 000002_init_accounts_and_categories.up.sql
CREATE TYPE account_type_enum AS ENUM ('CASH', 'BANK', 'E_WALLET', 'CREDIT_CARD');
CREATE TYPE category_type_enum AS ENUM ('INCOME', 'EXPENSE');

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type account_type_enum NOT NULL,
    balance BIGINT DEFAULT 0, -- Dalam satuan terkecil (Rupiah)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type category_type_enum NOT NULL,
    icon VARCHAR(50), -- Emoji atau identifier icon
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
