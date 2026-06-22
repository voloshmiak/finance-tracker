package server

import (
	"context"
	"errors"
	_ "finance-tracker/src/docs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

const shutdownTimeout = 10 * time.Second

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

func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Println("Starting server on port", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Println("Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	}
}
