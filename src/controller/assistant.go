package controller

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"finance-tracker/src/assistant"
	"finance-tracker/src/usecase"
)

type AssistantController struct {
	useCase  *usecase.Assistant
	username string
	password string
}

func NewAssistantController(useCase *usecase.Assistant, username, password string) *AssistantController {
	return &AssistantController{
		useCase:  useCase,
		username: username,
		password: password,
	}
}

// page godoc
// @Summary     Assistant chat UI
// @Description Serves the HTML chat interface for the AI finance assistant. Protected by HTTP Basic Auth.
// @Tags        Assistant
// @Produce     html
// @Success     200  {string}  string  "HTML page"
// @Failure     401  {string}  string  "Unauthorized"
// @Router      /assistant [get]
func (c *AssistantController) page(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = rw.Write([]byte(assistantHTML))
}

type askRequest struct {
	Question string `json:"question"`
	Model    string `json:"model"` // optional; empty uses the default model
}

type askResponse struct {
	Answer string `json:"answer"`
}

// ask godoc
// @Summary     Ask the assistant
// @Description Sends a natural-language question about your finances and returns the assistant's answer. Protected by HTTP Basic Auth.
// @Tags        Assistant
// @Accept      json
// @Produce     json
// @Param       body  body      askRequest   true  "Question payload"
// @Success     200   {object}  askResponse
// @Failure     400   {string}  string  "Invalid request"
// @Failure     401   {string}  string  "Unauthorized"
// @Failure     500   {string}  string  "Internal Server Error"
// @Router      /assistant/ask [post]
func (c *AssistantController) ask(rw http.ResponseWriter, r *http.Request) {
	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(rw, "invalid request", http.StatusBadRequest)
		return
	}

	answer, err := c.useCase.Ask(r.Context(), req.Question, req.Model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(askResponse{Answer: answer})
}

type modelsResponse struct {
	Default string            `json:"default"`
	Models  []assistant.Model `json:"models"`
}

// models godoc
// @Summary     List available models
// @Description Returns the Gemini models that support generation, plus the default. Protected by HTTP Basic Auth.
// @Tags        Assistant
// @Produce     json
// @Success     200  {object}  modelsResponse
// @Failure     401  {string}  string  "Unauthorized"
// @Failure     500  {string}  string  "Internal Server Error"
// @Router      /assistant/models [get]
func (c *AssistantController) models(rw http.ResponseWriter, r *http.Request) {
	models, err := c.useCase.Models(r.Context())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(modelsResponse{
		Default: c.useCase.DefaultModel(),
		Models:  models,
	})
}

// basicAuth wraps a handler with HTTP Basic Auth, comparing credentials in
// constant time to avoid leaking them via timing.
func (c *AssistantController) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(c.username)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(c.password)) == 1
		if !ok || !userMatch || !passMatch {
			rw.Header().Set("WWW-Authenticate", `Basic realm="Finance Assistant"`)
			http.Error(rw, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(rw, r)
	}
}

func (c *AssistantController) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /assistant", c.basicAuth(c.page))
	mux.HandleFunc("GET /assistant/models", c.basicAuth(c.models))
	mux.HandleFunc("POST /assistant/ask", c.basicAuth(c.ask))
}
