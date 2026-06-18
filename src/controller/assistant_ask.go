package controller

import (
	"encoding/json"
	"finance-tracker/src/server"
	"finance-tracker/src/usecase"
	"net/http"
)

type AssistantAsk struct {
	useCase *usecase.Assistant
}

func NewAssistantAsk(useCase *usecase.Assistant) *AssistantAsk {
	return &AssistantAsk{
		useCase: useCase,
	}
}

type askRequest struct {
	Question string `json:"question"`
	Model    string `json:"model"`
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
func (c *AssistantAsk) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		jsonError(rw, "invalid request", http.StatusBadRequest)
		return
	}

	answer, err := c.useCase.Ask(r.Context(), req.Question, req.Model)
	if err != nil {
		jsonError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(askResponse{Answer: answer})
}

func jsonError(rw http.ResponseWriter, msg string, code int) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	_ = json.NewEncoder(rw).Encode(map[string]string{"error": msg})
}

func (c *AssistantAsk) Route() server.Route {
	return server.Route{
		Pattern: "POST /assistant/ask",
		Handler: c,
		IsPage:  true,
	}
}
