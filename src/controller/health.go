package controller

import "net/http"

type HealthController struct {
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// ServeHTTP godoc
// @Summary     Health Check
// @Description Returns 200 OK when the service is healthy
// @Tags        System
// @Success     200  {string}  string  "OK"
// @Router      /health [get]
func (h *HealthController) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func (h *HealthController) Register(mux *http.ServeMux) {
	mux.Handle("GET /health", h)
}
