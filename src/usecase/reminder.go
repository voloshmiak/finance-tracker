package usecase

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
)

var ErrUnauthorized = errors.New("invalid auth token")

type Notifier interface {
	Notify(ctx context.Context, message string) error
}

type Reminder struct {
	notifier Notifier
	token    string
}

func NewReminder(notifier Notifier, token string) *Reminder {
	return &Reminder{
		notifier: notifier,
		token:    token,
	}
}

func (u *Reminder) Execute(ctx context.Context, auth string) error {
	const prefix = "Bearer "
	token := strings.TrimPrefix(auth, prefix)
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(u.token)) != 1 {
		return ErrUnauthorized
	}

	return u.notifier.Notify(ctx, "This is your reminder!")
}
