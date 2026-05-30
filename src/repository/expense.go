package repository

import (
	"context"
	"database/sql"
	"finance-tracker/src/entity"
)

type ExpenseRepository struct {
	conn *sql.DB
}

func NewExpenseRepository(conn *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{conn: conn}
}

func (r *ExpenseRepository) Create(ctx context.Context, expense *entity.Expense) error {
	query := `
		INSERT INTO expenses (id, amount, category, description, created_at) 
		VALUES ($1, $2, $3, $4, $5)
	`

	result, err := r.conn.ExecContext(ctx, query,
		expense.ID, expense.Amount, expense.Category,
		expense.Description, expense.CreatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
