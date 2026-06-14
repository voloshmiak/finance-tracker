package parser

import (
	"os"
	"testing"

	"finance-tracker/src/entity"
)

func TestCreditAgricoleParse(t *testing.T) {
	f, err := os.Open("/tmp/ca_test.xlsx")
	if err != nil {
		t.Skip("sample xlsx not present:", err)
	}
	defer f.Close()

	txs, err := NewCreditAgricole().Parse(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(txs) == 0 {
		t.Fatal("no transactions parsed")
	}

	tx := txs[0]
	t.Logf("parsed %d tx; first: %+v mask=%v", len(txs), tx, deref(tx.CardMask))

	if tx.Bank != entity.BankCreditAgricole {
		t.Errorf("bank = %q", tx.Bank)
	}
	if tx.Amount != -500000 { // -5000.0 UAH
		t.Errorf("amount = %d, want -500000", tx.Amount)
	}
	if tx.Currency != "UAH" {
		t.Errorf("currency = %q", tx.Currency)
	}
	if tx.MCC != nil {
		t.Errorf("MCC must be nil for CA, got %v", *tx.MCC)
	}
	if tx.OperatedAt.Year() != 2026 || tx.OperatedAt.Day() != 30 {
		t.Errorf("operatedAt = %v", tx.OperatedAt)
	}
	if tx.CardMask == nil {
		t.Error("card mask should be extracted from header")
	}
}

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
