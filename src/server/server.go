package server

import (
	"context"
	_ "finance-tracker/src/docs"
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	httpServer *http.Server
}

type Registrar interface {
	Register(mux *http.ServeMux)
}

func New(addr string, registrars ...Registrar) *Server {
	api := http.NewServeMux()
	for _, r := range registrars {
		r.Register(api)
	}
	root := http.NewServeMux()
	root.Handle("/api/", http.StripPrefix("/api", api))
	root.Handle("/swagger/", httpSwagger.WrapHandler)
	return &Server{
		httpServer: &http.Server{
			Addr:        addr,
			Handler:     root,
			ReadTimeout: 5 * time.Second,
			// WriteTimeout is generous because the assistant endpoint runs an
			// LLM tool-calling loop that can take well over 10s.
			WriteTimeout: 120 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	log.Println("Starting server on port", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
