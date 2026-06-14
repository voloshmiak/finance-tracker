CREATE TABLE transactions (
    id                uuid PRIMARY KEY,
    bank              text        NOT NULL,
    operated_at       timestamp with time zone NOT NULL,
    description       text        NOT NULL,
    amount            bigint      NOT NULL,
    currency          text        NOT NULL,

    mcc               integer,
    original_amount   bigint,
    original_currency text,
    commission        bigint,
    cashback          bigint,
    balance_after     bigint,
    card_mask         text,

    dedup_hash        text        NOT NULL,
    created_at        timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (dedup_hash)
);

CREATE INDEX idx_transactions_operated_at ON transactions (operated_at);
CREATE INDEX idx_transactions_bank ON transactions (bank);