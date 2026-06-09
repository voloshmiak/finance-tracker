CREATE TABLE cards (
    id UUID PRIMARY KEY,
    last4 CHAR(4) NOT NULL,
    bank TEXT NOT NULL,
    label TEXT,
    currency TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (last4, bank)
);