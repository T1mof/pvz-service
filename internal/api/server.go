package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"log/slog"
	"pvz-service/internal/config"
	"pvz-service/internal/logger"
)

type Server struct {
	server *http.Server
	log    *slog.Logger
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	log := slog.Default()

	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
			Handler:      handler,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: log,
	}
}

func (s *Server) Start() error {
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		s.log.Info("сервер завершает работу...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		s.server.SetKeepAlivesEnabled(false)
		if err := s.server.Shutdown(ctx); err != nil {
			s.log.Error("ошибка при корректном завершении сервера",
				"error", err,
				"timeout", "30s",
			)
			os.Exit(1)
		}
		close(done)
	}()

	s.log.Info("сервер запускается",
		"address", s.server.Addr,
		"read_timeout", s.server.ReadTimeout.String(),
		"write_timeout", s.server.WriteTimeout.String(),
		"idle_timeout", s.server.IdleTimeout.String(),
	)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error("ошибка запуска сервера", "error", err)
		return err
	}

	<-done
	s.log.Info("сервер корректно остановлен")
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) SetLogger(log *slog.Logger) {
	if log != nil {
		s.log = log
	}
}

func (s *Server) WithLogger(ctx context.Context) *Server {
	log := logger.FromContext(ctx)
	if log == nil {
		return s
	}

	serverCopy := *s
	serverCopy.log = log
	return &serverCopy
}
