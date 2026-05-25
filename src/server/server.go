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
	mux := http.NewServeMux()
	for _, r := range registrars {
		r.Register(mux)
	}
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
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
