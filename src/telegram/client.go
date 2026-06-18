package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

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

type getFileResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		FilePath string `json:"file_path"`
	} `json:"result"`
}

// DownloadFile resolves a file_id to its content. The caller must close the
// returned reader. It performs the two-step Telegram flow: getFile to learn the
// file path, then a download from the file endpoint.
func (c *Client) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	getURL := c.host + "/bot" + c.token + "/getFile?file_id=" + url.QueryEscape(fileID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create getFile request")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call getFile")
	}
	defer resp.Body.Close()

	var gf getFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&gf); err != nil {
		return nil, errors.Wrap(err, "failed to decode getFile response")
	}
	if !gf.Ok || gf.Result.FilePath == "" {
		return nil, errors.Wrap(ErrAPIError, "getFile returned no file path")
	}

	// File content lives under /file/bot<token>/<file_path>, not the API path.
	downloadURL := c.host + "/file/bot" + c.token + "/" + gf.Result.FilePath
	dreq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create download request")
	}
	dresp, err := c.httpClient.Do(dreq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download file")
	}
	if dresp.StatusCode != http.StatusOK {
		dresp.Body.Close()
		return nil, errors.Wrapf(ErrAPIError, "download failed with status %d", dresp.StatusCode)
	}

	return dresp.Body, nil
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

	var decodedRespBody sendMessageResponse
	if err = json.NewDecoder(resp.Body).Decode(&decodedRespBody); err != nil {
		return errors.Wrap(err, "failed to decode response")
	}

	if !decodedRespBody.Ok {
		return errors.Wrap(ErrAPIError, "telegram API returned not ok")
	}

	return nil
}
