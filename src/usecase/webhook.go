package usecase

import (
	"context"
	"finance-tracker/src/entity"
	"finance-tracker/src/parser"
	"finance-tracker/src/telegram"
	"fmt"
	"io"
	"regexp"
)

// last4Re extracts the 4-digit card identifier from a message caption.
var last4Re = regexp.MustCompile(`\b(\d{4})\b`)

// TransactionStore persists parsed transactions, skipping duplicates.
type TransactionStore interface {
	Save(ctx context.Context, txs []*entity.Transaction) (int, error)
}

// FileDownloader fetches the content of a Telegram file by its file_id.
type FileDownloader interface {
	DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error)
}

type Webhook struct {
	registry   *parser.Registry
	store      TransactionStore
	downloader FileDownloader
	notifier   Notifier
	chatID     int
}

func NewWebhook(registry *parser.Registry, store TransactionStore, downloader FileDownloader, notifier Notifier, chatID int) *Webhook {
	return &Webhook{
		registry:   registry,
		store:      store,
		downloader: downloader,
		notifier:   notifier,
		chatID:     chatID,
	}
}

// HandleUpdate processes a Telegram update: if it carries a document, the file
// is downloaded and imported. Updates from other chats or without a document
// are ignored (returning nil) so Telegram does not retry the webhook.
func (u *Webhook) HandleUpdate(ctx context.Context, update telegram.Update) error {
	msg := update.Message
	if msg == nil || msg.Document == nil {
		return nil
	}
	if msg.Chat.ID != u.chatID {
		return nil // ignore everyone but the owner
	}

	// The user sends the card's last 4 digits in the caption so we know which
	// card the statement belongs to (monobank's CSV carries no card number).
	m := last4Re.FindStringSubmatch(msg.Caption)
	if m == nil {
		_ = u.notifier.Notify(ctx, "⚠️ Додайте до файлу підпис із 4 цифрами картки, наприклад: 5617")
		return nil
	}
	last4 := m[1]

	doc := msg.Document
	file, err := u.downloader.DownloadFile(ctx, doc.FileID)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	defer file.Close()

	if _, err := u.Import(ctx, doc.FileName, last4, file); err != nil {
		// Report parse/format problems back to the user instead of failing silently.
		_ = u.notifier.Notify(ctx, fmt.Sprintf("⚠️ Не вдалося обробити %q: %v", doc.FileName, err))
		return err
	}
	return nil
}

// Import parses a bank statement (format inferred from filename), tags every
// transaction with the card's last 4 digits, stores them, and reports the
// result back over Telegram. It returns the number of newly inserted
// (non-duplicate) transactions.
func (u *Webhook) Import(ctx context.Context, filename, last4 string, r io.Reader) (int, error) {
	format, err := parser.FormatFromFilename(filename)
	if err != nil {
		return 0, err
	}

	p, err := u.registry.For(format)
	if err != nil {
		return 0, err
	}

	txs, err := p.Parse(r)
	if err != nil {
		return 0, fmt.Errorf("parse statement: %w", err)
	}

	for _, t := range txs {
		t.Last4 = last4
	}

	inserted, err := u.store.Save(ctx, txs)
	if err != nil {
		return 0, fmt.Errorf("save transactions: %w", err)
	}

	msg := fmt.Sprintf("✅ Звіт %q оброблено: %d нових операцій (усього в файлі %d).",
		filename, inserted, len(txs))
	if err := u.notifier.Notify(ctx, msg); err != nil {
		// The import succeeded; a failed notification shouldn't fail the request.
		return inserted, nil
	}

	return inserted, nil
}
