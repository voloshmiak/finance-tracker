package controller

import (
	"encoding/json"
	"finance-tracker/src/server"
	"finance-tracker/src/usecase"
	"net/http"
)

type AskAssistant struct {
	useCase *usecase.AskAssistant
}

func NewAskAssistant(useCase *usecase.AskAssistant) *AskAssistant {
	return &AskAssistant{
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
// @Tags        AskAssistant
// @Accept      json
// @Produce     json
// @Param       body  body      askRequest   true  "Question payload"
// @Success     200   {object}  askResponse
// @Failure     400   {string}  string  "Invalid request"
// @Failure     401   {string}  string  "Unauthorized"
// @Failure     500   {string}  string  "Internal Server Error"
// @Router      /assistant/ask [post]
func (c *AskAssistant) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		return
	}

	answer, err := c.useCase.Execute(r.Context(), req.Question, req.Model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(rw).Encode(
		askResponse{Answer: answer},
	); err != nil {
		http.Error(rw, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (c *AskAssistant) Route() server.Route {
	return server.Route{
		Pattern: "POST /assistant/ask",
		Handler: c,
		IsPage:  true,
	}
}
