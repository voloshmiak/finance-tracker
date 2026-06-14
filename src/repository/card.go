package repository

import (
	"context"
	"database/sql"
	"time"

	"finance-tracker/src/entity"
)

type CardRepository struct {
	conn *sql.DB
}

func NewCardRepository(conn *sql.DB) *CardRepository {
	return &CardRepository{conn: conn}
}

// MissingMonths returns, for each active card, the months since startFrom for
// which no transaction has been imported yet. Months are returned as the first
// day of the month. Cards with no gaps are omitted.
func (r *CardRepository) MissingMonths(ctx context.Context, startFrom time.Time) ([]entity.MissingReport, error) {
	const query = `
		WITH months AS (
			SELECT generate_series(
				date_trunc('month', $1::timestamptz),
				-- up to the previous month: the current month isn't over yet,
				-- so its statement can't exist.
				date_trunc('month', now()) - interval '1 month',
				interval '1 month'
			) AS month
		)
		SELECT c.last4, c.bank, c.label, m.month
		FROM cards c
		CROSS JOIN months m
		LEFT JOIN transactions t
			ON t.last4 = c.last4
			AND t.bank = c.bank
			AND date_trunc('month', t.operated_at) = m.month
		WHERE c.is_active
			AND t.id IS NULL
		GROUP BY c.last4, c.bank, c.label, m.month
		ORDER BY c.bank, c.last4, m.month
	`

	rows, err := r.conn.QueryContext(ctx, query, startFrom)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []entity.MissingReport
	for rows.Next() {
		var (
			rep   entity.MissingReport
			label sql.NullString
		)
		if err := rows.Scan(&rep.Last4, &rep.Bank, &label, &rep.Month); err != nil {
			return nil, err
		}
		if label.Valid {
			rep.Label = label.String
		}
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}
