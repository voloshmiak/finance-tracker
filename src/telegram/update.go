package telegram

// Update is a single incoming update delivered to the webhook.
// https://core.telegram.org/bots/api#update
type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID int       `json:"message_id"`
	Chat      Chat      `json:"chat"`
	Document  *Document `json:"document"`
	Caption   string    `json:"caption"` // text sent alongside a file (e.g. card last 4 digits)
}

type Chat struct {
	ID int `json:"id"`
}

// Document describes a file attached to a message (a bank statement upload).
type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
}
