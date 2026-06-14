package parser

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"finance-tracker/src/entity"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// creditAgricoleLayout is the date format CA uses: "30.05.2026" (no time).
const creditAgricoleLayout = "02.01.2006"

// credit agricole operations table column indexes (0-based, as returned by GetRows).
const (
	caColNum         = 0  // № Карт.
	caColOperatedAt  = 1  // Дата здійснення операції
	caColDescription = 6  // Опис операції
	caColCurrency    = 8  // Валюта операції
	caColOrigAmount  = 9  // Сума у валюті операції
	caColAmount      = 10 // Сума у валюті рахунку
)

// cardMaskRe matches a masked card number like "**** **** **** 5617".
var cardMaskRe = regexp.MustCompile(`\*+[\d* ]*\d{4}`)

// CreditAgricole parses Credit Agricole's XLSX statement export.
type CreditAgricole struct{}

func NewCreditAgricole() *CreditAgricole {
	return &CreditAgricole{}
}

func (p *CreditAgricole) Parse(r io.Reader) ([]*entity.Transaction, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets in workbook")
	}

	// RawCellValue avoids excelize applying the cell's display format, which
	// would turn "-5000.0" into a thousands-separated "-5.000.00".
	rows, err := f.GetRows(sheets[0], excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}

	cardMask := extractCardMask(rows)

	var txs []*entity.Transaction
	inOperations := false
	for _, row := range rows {
		a := strings.TrimSpace(cell(row, caColNum))

		// Section boundaries.
		if strings.HasPrefix(a, "Операції за рахунком") {
			inOperations = true
			continue
		}
		if !inOperations {
			continue
		}
		// The operations section ends at the blockings section or summary rows.
		if strings.HasPrefix(a, "Блокування") {
			break
		}
		desc := strings.TrimSpace(cell(row, caColDescription))
		if strings.HasPrefix(desc, "Всього") {
			break
		}

		// A data row starts with a numeric "№ Карт." (e.g. "1.0"); skip the header.
		if a == "" || a == "№ Карт." || !startsWithDigit(a) {
			continue
		}

		tx, err := p.parseRow(row, cardMask)
		if err != nil {
			return nil, fmt.Errorf("row %q: %w", a, err)
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

func (p *CreditAgricole) parseRow(row []string, cardMask *string) (*entity.Transaction, error) {
	rawDate := strings.TrimSpace(cell(row, caColOperatedAt))
	operatedAt, err := time.ParseInLocation(creditAgricoleLayout, rawDate, kyivLocation)
	if err != nil {
		return nil, fmt.Errorf("parse date %q: %w", rawDate, err)
	}

	amount, _, err := parseMoneyKop(cell(row, caColAmount))
	if err != nil {
		return nil, fmt.Errorf("parse amount: %w", err)
	}

	tx := &entity.Transaction{
		ID:          uuid.New(),
		Bank:        entity.BankCreditAgricole,
		OperatedAt:  operatedAt,
		Description: strings.TrimSpace(cell(row, caColDescription)),
		Amount:      amount,
		Currency:    strings.TrimSpace(cell(row, caColCurrency)),
		CardMask:    cardMask,
		// MCC stays nil: CA provides "Код операції", not an MCC.
	}

	if v, ok, err := parseMoneyKop(cell(row, caColOrigAmount)); err != nil {
		return nil, fmt.Errorf("parse original amount: %w", err)
	} else if ok {
		tx.OriginalAmount = ptr(v)
	}

	return tx, nil
}

// cell safely reads column i from a row that GetRows may have trimmed.
func cell(row []string, i int) string {
	if i < len(row) {
		return row[i]
	}
	return ""
}

func startsWithDigit(s string) bool {
	return s != "" && s[0] >= '0' && s[0] <= '9'
}

// extractCardMask finds the masked card number from the statement header.
func extractCardMask(rows [][]string) *string {
	for _, row := range rows {
		for _, c := range row {
			if mask := cardMaskRe.FindString(c); mask != "" {
				m := strings.TrimSpace(mask)
				return &m
			}
		}
	}
	return nil
}
