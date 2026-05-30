package controller

import (
	"context"
	"encoding/json"
	"finance-tracker/src/entity"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ExpenseRepository interface {
	Create(ctx context.Context, expense *entity.Expense) error
}

type CreateExpenseController struct {
	repo ExpenseRepository
}

func NewCreateExpenseController(repo ExpenseRepository) *CreateExpenseController {
	return &CreateExpenseController{
		repo: repo,
	}
}

type CreateExpenseRequest struct {
	Amount      int    `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

func (r *CreateExpenseRequest) toExpense() *entity.Expense {
	return &entity.Expense{
		ID:          uuid.New(),
		Amount:      r.Amount,
		Category:    r.Category,
		Description: r.Description,
		CreatedAt:   time.Now(),
	}
}

// ServeHTTP godoc
// @Summary     Create Expense
// @Description Adds a new expense entry to the finance tracker
// @Tags        Expense
// @Accept      json
// @Produce     json
// @Param       body  body      CreateExpenseRequest  true  "Expense payload"
// @Success     201   {object}  entity.Expense
// @Failure     400   {object}  string  "Invalid request body"
// @Failure     500   {object}  string  "Failed to create expense"
// @Router      /expenses [post]
func (c *CreateExpenseController) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var body CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err := c.repo.Create(r.Context(), body.toExpense())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(body.toExpense()); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusCreated)
}

func (c *CreateExpenseController) Register(mux *http.ServeMux) {
	mux.Handle("POST /expenses", c)
}
