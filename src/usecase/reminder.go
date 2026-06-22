package usecase

import (
	"context"
	"crypto/subtle"
	"errors"
	"finance-tracker/src/entity"
	"fmt"
	"strings"
	"time"
)

var ErrUnauthorized = errors.New("invalid auth token")

// staleAfter is how long a card's coverage may lag before we remind about it.
const staleAfter = 7 * 24 * time.Hour

type Notifier interface {
	Notify(ctx context.Context, message string) error
}

// UncoveredStore reports cards whose statements are not imported up to date.
type UncoveredStore interface {
	Uncovered(ctx context.Context, staleBefore time.Time) ([]entity.UncoveredCard, error)
}

type Reminder struct {
	notifier Notifier
	cards    UncoveredStore
	token    string
}

func NewReminder(notifier Notifier, cards UncoveredStore, token string) *Reminder {
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

	now := time.Now()
	cards, err := u.cards.Uncovered(ctx, now.Add(-staleAfter))
	if err != nil {
		return fmt.Errorf("query uncovered cards: %w", err)
	}

	return u.notifier.Notify(ctx, buildReminder(cards, now))
}

// buildReminder lists each card with the date range still missing.
func buildReminder(cards []entity.UncoveredCard, now time.Time) string {
	if len(cards) == 0 {
		return "✅ Усі звіти актуальні — нічого довантажувати."
	}

	today := now.Format("02.01.2006")
	var b strings.Builder
	b.WriteString("⏰ Час довантажити звіти. Не вистачає за періоди:\n")
	for _, c := range cards {
		title := fmt.Sprintf("%s ****%s", c.Bank, c.Last4)
		if c.Label != "" {
			title = fmt.Sprintf("%s (%s ****%s)", c.Label, c.Bank, c.Last4)
		}

		if c.CoveredUntil == nil {
			fmt.Fprintf(&b, "\n• %s:\n  жодного звіту ще не завантажено", title)
			continue
		}
		from := c.CoveredUntil.Add(24 * time.Hour).Format("02.01.2006")
		fmt.Fprintf(&b, "\n• %s:\n  з %s по %s", title, from, today)
	}
	return b.String()
}
