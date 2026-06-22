-- covered_until marks how far statements have been imported for a card.
-- It is advanced on each import to the latest transaction date seen.
ALTER TABLE cards ADD COLUMN covered_until timestamptz;

-- Seed from existing data so the first reminder is meaningful.
UPDATE cards c
SET covered_until = sub.max_date
FROM (
    SELECT last4, bank, MAX(operated_at) AS max_date
    FROM transactions
    GROUP BY last4, bank
) sub
WHERE c.last4 = sub.last4 AND c.bank = sub.bank;