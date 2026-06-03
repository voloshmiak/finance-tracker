package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

var ErrAPIError = errors.New("telegram API error")

type Client struct {
	token      string
	host       string
	chatID     int
	httpClient *http.Client
}

func NewClient(token, host string, chatID int) *Client {
	return &Client{
		token:      token,
		host:       host,
		chatID:     chatID,
		httpClient: http.DefaultClient,
	}
}

type sendMessageRequest struct {
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}

type sendMessageResponse struct {
	Ok     bool `json:"ok"`
	Result any  `json:"result"`
}

type errorResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

func (c *Client) Notify(ctx context.Context, message string) error {
	url := c.host + "/bot" + c.token + "/sendMessage"
	reqBody := sendMessageRequest{
		ChatID: c.chatID,
		Text:   message,
	}
	encodedReqBody, err := json.Marshal(reqBody)
	if err != nil {
		return errors.Wrap(err, "failed to encode message")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(encodedReqBody))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err = json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return err
		}
		return errors.Wrap(ErrAPIError, errResp.Description)
	}

	var encodedRespBody sendMessageResponse
	if err = json.NewDecoder(resp.Body).Decode(&encodedRespBody); err != nil {
		return errors.Wrap(err, "failed to decode response")
	}

	if !encodedRespBody.Ok {
		return errors.Wrap(ErrAPIError, "telegram API returned not ok")
	}

	return nil
}
