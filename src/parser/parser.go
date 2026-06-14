package parser

import (
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"finance-tracker/src/entity"
)

// ErrUnsupportedFormat is returned when a statement format has no parser.
var ErrUnsupportedFormat = errors.New("unsupported statement format")

type BankParser interface {
	Parse(r io.Reader) ([]*entity.Transaction, error)
}

// kyivLocation is the timezone all Ukrainian bank statements are expressed in.
// Statements carry local Kyiv time without an explicit offset, so we attach it
// when parsing to keep stored timestamps correct.
var kyivLocation = mustLoadKyiv()

func mustLoadKyiv() *time.Location {
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		return time.FixedZone("EET", 2*60*60)
	}
	return loc
}

const emDash = "—"

func isBlank(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == emDash || s == "-"
}

func parseMoneyKop(s string) (int, bool, error) {
	if isBlank(s) {
		return 0, false, nil
	}
	clean := strings.ReplaceAll(strings.TrimSpace(s), " ", "")
	clean = strings.ReplaceAll(clean, " ", "") // non-breaking space
	clean = strings.ReplaceAll(clean, ",", ".")

	f, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, false, err
	}
	return int(math.Round(f * 100)), true, nil
}

func ptr[T any](v T) *T {
	return &v
}
