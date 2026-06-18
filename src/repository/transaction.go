package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"finance-tracker/src/entity"
	"fmt"
)

type Transactions struct {
	conn *sql.DB
}

func NewTransactions(conn *sql.DB) *Transactions {
	return &Transactions{conn: conn}
}

func dedupHash(t *entity.Transaction) string {
	raw := fmt.Sprintf("%s|%s|%s|%d|%s",
		t.Bank, t.Last4, t.OperatedAt.UTC().Format("2006-01-02T15:04:05"), t.Amount, t.Description)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (r *Transactions) Save(ctx context.Context, txs []*entity.Transaction) (int, error) {
	if len(txs) == 0 {
		return 0, nil
	}

	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO transactions (
			id, bank, operated_at, description, amount, currency,
			mcc, original_amount, original_currency, commission, cashback,
			balance_after, card_mask, last4, dedup_hash
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (dedup_hash) DO NOTHING
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	inserted := 0
	for _, t := range txs {
		res, err := stmt.ExecContext(ctx,
			t.ID, t.Bank, t.OperatedAt, t.Description, t.Amount, t.Currency,
			t.MCC, t.OriginalAmount, t.OriginalCurrency, t.Commission, t.Cashback,
			t.BalanceAfter, t.CardMask, t.Last4, dedupHash(t),
		)
		if err != nil {
			return 0, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return inserted, nil
}
