package server

import (
	"context"
	_ "finance-tracker/src/docs"
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Controller interface {
	Route() Route
}

type Server struct {
	httpServer *http.Server
	api        *http.ServeMux
	auth       Middleware
}

type Route struct {
	Pattern string
	Handler http.Handler
	IsPage  bool
}

func New(addr string, auth Middleware, global ...Middleware) *Server {
	api := http.NewServeMux()
	root := http.NewServeMux()
	root.Handle("/api/", http.StripPrefix("/api", api))
	root.Handle("/swagger/", httpSwagger.WrapHandler)

	return &Server{
		api: api,
		httpServer: &http.Server{
			Addr:         ":" + addr,
			Handler:      root,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 120 * time.Second,
		},
		auth: auth,
	}
}

func (s *Server) Register(controllers ...Controller) {
	for _, c := range controllers {
		r := c.Route()
		h := r.Handler
		if r.IsPage {
			h = s.auth(h)
		}
		s.api.Handle(r.Pattern, h)
	}
}

func (s *Server) Start() error {
	log.Println("Starting server on port", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
