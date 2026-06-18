package parser

import (
	"finance-tracker/src/entity"
	"os"
	"testing"
)

func TestAbankParse(t *testing.T) {
	f, err := os.Open("/tmp/abank_test.pdf")
	if err != nil {
		t.Skip("sample pdf not present:", err)
	}
	defer f.Close()

	txs, err := NewAbank().Parse(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(txs) == 0 {
		t.Fatal("no transactions parsed")
	}
	t.Logf("parsed %d transactions", len(txs))

	first := txs[0]
	t.Logf("first: %+v", first)

	if first.Bank != entity.BankAbank {
		t.Errorf("bank = %q", first.Bank)
	}
	if first.Description != "Novus" {
		t.Errorf("description = %q, want Novus", first.Description)
	}
	if first.Amount != -19600 { // -196.00 UAH
		t.Errorf("amount = %d, want -19600", first.Amount)
	}
	if first.Currency != "UAH" {
		t.Errorf("currency = %q", first.Currency)
	}
	if first.MCC == nil || *first.MCC != 5411 {
		t.Errorf("mcc = %v, want 5411", first.MCC)
	}
	if first.CardMask == nil || *first.CardMask != "4323****9061" {
		t.Errorf("cardMask = %v", first.CardMask)
	}
	if first.Cashback == nil || *first.Cashback != 137 { // 1.37
		t.Errorf("cashback = %v, want 137", first.Cashback)
	}
	if first.BalanceAfter == nil || *first.BalanceAfter != 6526298 { // 65 262.98
		t.Errorf("balanceAfter = %v, want 6526298", first.BalanceAfter)
	}
	if first.OperatedAt.Day() != 8 || first.OperatedAt.Hour() != 19 || first.OperatedAt.Minute() != 48 {
		t.Errorf("operatedAt = %v", first.OperatedAt)
	}

	// "Київ Цифровий" — multi-word description must stay intact
	foundMultiWord := false
	for _, tx := range txs {
		if tx.Description == "Київ Цифровий" {
			foundMultiWord = true
			break
		}
	}
	if !foundMultiWord {
		t.Error("multi-word description 'Київ Цифровий' not found")
	}
}
