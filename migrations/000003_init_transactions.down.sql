-- 000003_init_transactions.down.sql
DROP INDEX IF EXISTS idx_accounts_user;
DROP INDEX IF EXISTS idx_transactions_user_date;
DROP TABLE IF EXISTS transactions;
DROP TYPE IF EXISTS tx_type_enum;
