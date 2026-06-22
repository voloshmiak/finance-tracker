INSERT INTO mcc_categories (mcc, category, title, is_spending) VALUES
    (7999, 'Розваги', 'Рекреаційні послуги (квести, атракціони, активності)', true)
ON CONFLICT (mcc) DO NOTHING;