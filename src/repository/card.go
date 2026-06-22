package repository

import (
	"context"
	"database/sql"
	"finance-tracker/src/entity"
	"time"
)

type Cards struct {
	conn *sql.DB
}

func NewCards(conn *sql.DB) *Cards {
	return &Cards{conn: conn}
}

// Uncovered returns active cards whose statements are not imported up to a
// recent date: covered_until is older than `staleBefore`, or never set.
// Sorted by oldest coverage first.
func (r *Cards) Uncovered(ctx context.Context, staleBefore time.Time) ([]entity.UncoveredCard, error) {
	const query = `
		SELECT last4, bank, label, covered_until
		FROM cards
		WHERE is_active
			AND (covered_until IS NULL OR covered_until < $1)
		ORDER BY covered_until ASC NULLS FIRST
	`

	rows, err := r.conn.QueryContext(ctx, query, staleBefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []entity.UncoveredCard
	for rows.Next() {
		var (
			c            entity.UncoveredCard
			label        sql.NullString
			coveredUntil sql.NullTime
		)
		if err := rows.Scan(&c.Last4, &c.Bank, &label, &coveredUntil); err != nil {
			return nil, err
		}
		if label.Valid {
			c.Label = label.String
		}
		if coveredUntil.Valid {
			c.CoveredUntil = &coveredUntil.Time
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}
