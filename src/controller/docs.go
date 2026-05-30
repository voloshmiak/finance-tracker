package controller

import (
	"net/http"
)

type DocsController struct {
}

func NewDocsController() *DocsController {
	return &DocsController{}
}

// ServeHTTP godoc
// @Summary     API Documentation
// @Description Serves the interactive API reference UI
// @Tags        System
// @Produce     html
// @Success     200  {string}  string  "HTML page"
// @Router      /docs [get]
func (h *DocsController) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	html := `
    <!DOCTYPE html>
    <html>
      <head>
        <title>Finance Tracker API</title>
        <meta charset="utf-8" />
        <style>body { margin: 0; }</style>
      </head>
      <body>
        <script id="api-reference" data-url="/swagger/doc.json"></script>
        <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
      </body>
    </html>
    `
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := rw.Write([]byte(html))
	if err != nil {
		http.Error(rw, "Failed to load API reference", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *DocsController) Register(mux *http.ServeMux) {
	mux.Handle("GET /docs", h)
}
