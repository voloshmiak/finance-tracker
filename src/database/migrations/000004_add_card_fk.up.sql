-- Ensure every (last4, bank) already present in transactions has a card,
-- otherwise adding the foreign key would fail on existing rows.
INSERT INTO cards (id, last4, bank)
SELECT gen_random_uuid(), last4, bank
FROM transactions
ON CONFLICT (last4, bank) DO NOTHING;

ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_card
    FOREIGN KEY (last4, bank) REFERENCES cards (last4, bank);