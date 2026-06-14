-- Flag categories that move money rather than spend it (transfers, ATM cash),
-- so they can be excluded from spending analytics.
ALTER TABLE mcc_categories ADD COLUMN is_spending boolean NOT NULL DEFAULT true;

UPDATE mcc_categories SET is_spending = false
WHERE category IN ('Перекази', 'Готівка', 'Фінанси');