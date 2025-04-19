package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"pvz-service/internal/config"
)

type Server struct {
	server *http.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
			Handler:      handler,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		s.server.SetKeepAlivesEnabled(false)
		if err := s.server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v", err)
		}
		close(done)
	}()

	log.Printf("Server is starting on %s", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-done
	log.Println("Server stopped")
	return nil
}

// Метод Shutdown для graceful shutdown сервера
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
