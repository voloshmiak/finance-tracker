package controller

import (
	"finance-tracker/src/server"
	"net/http"
)

var html = `
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

type Docs struct{}

func NewDocs() *Docs {
	return &Docs{}
}

// ServeHTTP godoc
// @Summary     API Documentation
// @Description Serves the interactive API reference UI
// @Tags        System
// @Produce     html
// @Success     200  {string}  string  "HTML page"
// @Router      /docs [get]
func (h *Docs) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := rw.Write([]byte(html)); err != nil {
		http.Error(rw, "Failed to load API reference", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *Docs) Route() server.Route {
	return server.Route{
		Pattern: "GET /docs",
		Handler: h,
		IsPage:  true,
	}
}
