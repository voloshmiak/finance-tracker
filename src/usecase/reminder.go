package usecase

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"time"

	"finance-tracker/src/entity"
)

var ErrUnauthorized = errors.New("invalid auth token")

// reportsSince is the earliest month we expect statements for.
var reportsSince = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

type Notifier interface {
	Notify(ctx context.Context, message string) error
}

// MissingReportStore reports which (card, month) statements are not yet imported.
type MissingReportStore interface {
	MissingMonths(ctx context.Context, startFrom time.Time) ([]entity.MissingReport, error)
}

type Reminder struct {
	notifier Notifier
	cards    MissingReportStore
	token    string
}

func NewReminder(notifier Notifier, cards MissingReportStore, token string) *Reminder {
	return &Reminder{
		notifier: notifier,
		cards:    cards,
		token:    token,
	}
}

func (u *Reminder) Execute(ctx context.Context, auth string) error {
	const prefix = "Bearer "
	token := strings.TrimPrefix(auth, prefix)
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(u.token)) != 1 {
		return ErrUnauthorized
	}

	missing, err := u.cards.MissingMonths(ctx, reportsSince)
	if err != nil {
		return fmt.Errorf("query missing reports: %w", err)
	}

	return u.notifier.Notify(ctx, buildReminder(missing))
}

// buildReminder groups missing months per card into a readable Telegram message.
func buildReminder(missing []entity.MissingReport) string {
	if len(missing) == 0 {
		return "✅ Усі звіти завантажено — пропусків немає."
	}

	// Preserve order from the query (sorted by bank, last4, month) while grouping.
	type group struct {
		title  string
		months []string
	}
	var groups []*group
	index := make(map[string]*group)

	for _, m := range missing {
		key := m.Bank + "|" + m.Last4
		g, ok := index[key]
		if !ok {
			title := fmt.Sprintf("%s ****%s", m.Bank, m.Last4)
			if m.Label != "" {
				title = fmt.Sprintf("%s (%s ****%s)", m.Label, m.Bank, m.Last4)
			}
			g = &group{title: title}
			index[key] = g
			groups = append(groups, g)
		}
		g.months = append(g.months, m.Month.Format("2006-01"))
	}

	var b strings.Builder
	b.WriteString("⏰ Не вистачає звітів за такі періоди:\n")
	for _, g := range groups {
		fmt.Fprintf(&b, "\n• %s:\n  %s", g.title, strings.Join(g.months, ", "))
	}
	return b.String()
}
