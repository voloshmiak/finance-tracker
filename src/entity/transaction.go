package entity

import (
	"time"

	"github.com/google/uuid"
)

// MissingReport is a (card, month) pair for which no statement has been imported.
type MissingReport struct {
	Last4 string
	Bank  string
	Label string // human-friendly card name, empty if unset
	Month time.Time
}

type Bank string

const (
	BankMonobank       Bank = "monobank"
	BankCreditAgricole Bank = "credit_agricole"
	BankAbank          Bank = "abank"
)

type Transaction struct {
	ID          uuid.UUID `json:"id"`
	Bank        Bank      `json:"bank"`
	OperatedAt  time.Time `json:"operated_at"` // time is 00:00 when the bank gives only a date
	Description string    `json:"description"`
	Amount      int       `json:"amount"`   // in the account currency, kopiykas
	Currency    string    `json:"currency"` // account currency, e.g. "UAH"

	MCC              *int    `json:"mcc,omitempty"`               // monobank, abank
	OriginalAmount   *int    `json:"original_amount,omitempty"`   // amount in operation currency, kopiykas
	OriginalCurrency *string `json:"original_currency,omitempty"` // operation currency
	Commission       *int    `json:"commission,omitempty"`        // monobank, abank
	Cashback         *int    `json:"cashback,omitempty"`          // monobank, abank
	BalanceAfter     *int    `json:"balance_after,omitempty"`     // monobank, abank
	CardMask         *string `json:"card_mask,omitempty"`         // abank (per row), credit_agricole (header)

	Last4 string `json:"last4,omitempty"` // card last 4 digits, supplied by the user in the message caption

	DedupHash string    `json:"-"`          // natural key for idempotent imports
	CreatedAt time.Time `json:"created_at"` // when we stored the row
}
