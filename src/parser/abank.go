package parser

import (
	"bytes"
	"finance-tracker/src/entity"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

// abankLayout matches the merged date+time anchor: "08.05.202619:48".
const abankLayout = "02.01.200615:04"

// abankAnchorRe detects the first line of a transaction block (date+time glued).
var abankAnchorRe = regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}\d{2}:\d{2}$`)

// Lines per transaction block in a-bank's GetPlainText output:
//
//	0  08.05.202619:48   date+time
//	1  4323****9061      card mask
//	2  Novus             description (may contain spaces)
//	3  5411              MCC
//	4  -196.00           amount (card currency)
//	5  -196.00           amount (operation currency)
//	6  UAH               currency
//	7  -                 rate
//	8  0.00              commission
//	9  1.37              cashback
//	10 65 262.98         balance after
const abankBlockLen = 11

const (
	abColDateTime = iota
	abColCard
	abColDescription
	abColMCC
	abColAmount
	abColOrigAmount
	abColCurrency
	abColRate
	abColCommission
	abColCashback
	abColBalanceAfter
)

// Abank parses AKCENT-BANK (a-bank) PDF statement exports.
type Abank struct{}

func NewAbank() *Abank {
	return &Abank{}
}

func (p *Abank) Parse(r io.Reader) ([]*entity.Transaction, error) {
	// ledongthuc/pdf needs a ReaderAt + size, so buffer the stream.
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read pdf: %w", err)
	}

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open pdf: %w", err)
	}

	textReader, err := reader.GetPlainText()
	if err != nil {
		return nil, fmt.Errorf("extract text: %w", err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(textReader); err != nil {
		return nil, fmt.Errorf("buffer text: %w", err)
	}

	var lines []string
	for l := range strings.SplitSeq(buf.String(), "\n") {
		lines = append(lines, strings.TrimSpace(l))
	}

	var txs []*entity.Transaction
	for i := 0; i < len(lines); i++ {
		if !abankAnchorRe.MatchString(lines[i]) {
			continue
		}
		if i+abankBlockLen > len(lines) {
			break // truncated block at EOF
		}
		tx, err := p.parseBlock(lines[i : i+abankBlockLen])
		if err != nil {
			return nil, fmt.Errorf("block at line %d: %w", i, err)
		}
		txs = append(txs, tx)
		i += abankBlockLen - 1
	}

	return txs, nil
}

func (p *Abank) parseBlock(b []string) (*entity.Transaction, error) {
	operatedAt, err := time.ParseInLocation(abankLayout, b[abColDateTime], kyivLocation)
	if err != nil {
		return nil, fmt.Errorf("parse datetime %q: %w", b[abColDateTime], err)
	}

	amount, _, err := parseMoneyKop(b[abColAmount])
	if err != nil {
		return nil, fmt.Errorf("parse amount %q: %w", b[abColAmount], err)
	}

	tx := &entity.Transaction{
		ID:          uuid.New(),
		Bank:        entity.BankAbank,
		OperatedAt:  operatedAt,
		Description: b[abColDescription],
		Amount:      amount,
		Currency:    b[abColCurrency],
		CardMask:    new(b[abColCard]),
	}

	if !isBlank(b[abColMCC]) {
		mcc, err := strconv.Atoi(b[abColMCC])
		if err != nil {
			return nil, fmt.Errorf("parse mcc %q: %w", b[abColMCC], err)
		}
		tx.MCC = new(mcc)
	}

	if v, ok, err := parseMoneyKop(b[abColOrigAmount]); err != nil {
		return nil, fmt.Errorf("parse original amount: %w", err)
	} else if ok {
		tx.OriginalAmount = new(v)
	}

	if v, ok, err := parseMoneyKop(b[abColCommission]); err != nil {
		return nil, fmt.Errorf("parse commission: %w", err)
	} else if ok {
		tx.Commission = new(v)
	}

	if v, ok, err := parseMoneyKop(b[abColCashback]); err != nil {
		return nil, fmt.Errorf("parse cashback: %w", err)
	} else if ok {
		tx.Cashback = new(v)
	}

	if v, ok, err := parseMoneyKop(b[abColBalanceAfter]); err != nil {
		return nil, fmt.Errorf("parse balance: %w", err)
	} else if ok {
		tx.BalanceAfter = new(v)
	}

	return tx, nil
}
