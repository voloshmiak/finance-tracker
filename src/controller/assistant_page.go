package controller

import (
	_ "embed"
	"finance-tracker/src/server"
	"finance-tracker/src/usecase"
	"net/http"
)

//go:embed assistant.html
var assistantHTML string

type AssistantPage struct {
	useCase *usecase.Assistant
}

func NewAssistantPage(useCase *usecase.Assistant) *AssistantPage {
	return &AssistantPage{
		useCase: useCase,
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
func (c *AssistantPage) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := rw.Write([]byte(assistantHTML)); err != nil {
		http.Error(rw, "Failed to load assistant page", http.StatusInternalServerError)
		return
	}
}

func (c *AssistantPage) Route() server.Route {
	return server.Route{
		Pattern: "GET /assistant",
		Handler: c,
		IsPage:  true,
	}
}
