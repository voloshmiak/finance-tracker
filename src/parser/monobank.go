package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"finance-tracker/src/entity"

	"github.com/google/uuid"
)

const monobankLayout = "02.01.2006 15:04:05"

const (
	monoColDate = iota
	monoColDescription
	monoColMCC
	monoColAmount     // у валюті картки (UAH)
	monoColOrigAmount // у валюті операції
	monoColCurrency
	monoColRate
	monoColCommission
	monoColCashback
	monoColBalanceAfter
	monoColCount
)

type Monobank struct{}

func NewMonobank() *Monobank {
	return &Monobank{}
}

func (p *Monobank) Parse(r io.Reader) ([]*entity.Transaction, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = monoColCount
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(rows) < 2 {
		return nil, nil // header only or empty
	}

	txs := make([]*entity.Transaction, 0, len(rows)-1)
	for i, row := range rows[1:] { // skip header
		tx, err := p.parseRow(row)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (p *Monobank) parseRow(row []string) (*entity.Transaction, error) {
	operatedAt, err := time.ParseInLocation(monobankLayout, strings.TrimSpace(row[monoColDate]), kyivLocation)
	if err != nil {
		return nil, fmt.Errorf("parse date %q: %w", row[monoColDate], err)
	}

	amount, _, err := parseMoneyKop(row[monoColAmount])
	if err != nil {
		return nil, fmt.Errorf("parse amount %q: %w", row[monoColAmount], err)
	}

	tx := &entity.Transaction{
		ID:          uuid.New(),
		Bank:        entity.BankMonobank,
		OperatedAt:  operatedAt,
		Description: strings.TrimSpace(row[monoColDescription]),
		Amount:      amount,
		Currency:    strings.TrimSpace(row[monoColCurrency]),
	}

	if !isBlank(row[monoColMCC]) {
		mcc, err := strconv.Atoi(strings.TrimSpace(row[monoColMCC]))
		if err != nil {
			return nil, fmt.Errorf("parse mcc %q: %w", row[monoColMCC], err)
		}
		tx.MCC = ptr(mcc)
	}

	if v, ok, err := parseMoneyKop(row[monoColOrigAmount]); err != nil {
		return nil, fmt.Errorf("parse original amount: %w", err)
	} else if ok {
		tx.OriginalAmount = ptr(v)
	}

	if v, ok, err := parseMoneyKop(row[monoColCommission]); err != nil {
		return nil, fmt.Errorf("parse commission: %w", err)
	} else if ok {
		tx.Commission = ptr(v)
	}

	if v, ok, err := parseMoneyKop(row[monoColCashback]); err != nil {
		return nil, fmt.Errorf("parse cashback: %w", err)
	} else if ok {
		tx.Cashback = ptr(v)
	}

	if v, ok, err := parseMoneyKop(row[monoColBalanceAfter]); err != nil {
		return nil, fmt.Errorf("parse balance: %w", err)
	} else if ok {
		tx.BalanceAfter = ptr(v)
	}

	return tx, nil
}
