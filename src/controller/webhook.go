package controller

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"finance-tracker/src/telegram"
	"finance-tracker/src/usecase"
)

// secretTokenHeader is set by Telegram when the webhook is registered with a
// secret_token, letting us verify the request really comes from Telegram.
const secretTokenHeader = "X-Telegram-Bot-Api-Secret-Token"

type WebhookController struct {
	useCase *usecase.Webhook
	secret  string
}

func NewWebhookController(useCase *usecase.Webhook, secret string) *WebhookController {
	return &WebhookController{
		useCase: useCase,
		secret:  secret,
	}
}

// ServeHTTP godoc
// @Summary     Telegram webhook
// @Description Receives Telegram bot updates. When a message carries a bank statement document (monobank CSV, Credit Agricole XLSX, or a-bank PDF), the file is downloaded, parsed, deduplicated and stored. Authenticated via the Telegram secret token header.
// @Tags        Telegram
// @Accept      json
// @Param       X-Telegram-Bot-Api-Secret-Token  header  string           true  "Telegram webhook secret token"
// @Param       update                           body    telegram.Update  true  "Telegram update payload"
// @Success     200  "Update processed"
// @Failure     401  {object}  string  "Invalid secret token"
// @Failure     500  {object}  string  "Internal Server Error"
// @Router      /webhook [post]
func (c *WebhookController) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	got := r.Header.Get(secretTokenHeader)
	if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(c.secret)) != 1 {
		http.Error(rw, "unauthorized", http.StatusUnauthorized)
		return
	}

	var update telegram.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(rw, "invalid update payload", http.StatusBadRequest)
		return
	}

	if err := c.useCase.HandleUpdate(r.Context(), update); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (c *WebhookController) Register(mux *http.ServeMux) {
	mux.Handle("POST /webhook", c)
}
