DROP INDEX IF EXISTS idx_transactions_last4;

ALTER TABLE transactions DROP COLUMN IF EXISTS last4;