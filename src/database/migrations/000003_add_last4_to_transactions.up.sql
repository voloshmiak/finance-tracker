ALTER TABLE transactions ADD COLUMN last4 CHAR(4) NOT NULL DEFAULT '';

CREATE INDEX idx_transactions_last4 ON transactions (last4);