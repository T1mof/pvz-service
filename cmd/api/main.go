package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pvz-service/internal/api"
	"pvz-service/internal/api/middleware"
	"pvz-service/internal/config"
	"pvz-service/internal/grpc"
	"pvz-service/internal/logger"
	"pvz-service/internal/repository/postgres"
	"pvz-service/internal/services"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.New(logger.Config{
		Level:       logger.LevelInfo,
		Format:      "json",
		Output:      os.Stdout,
		ServiceName: "pvz-service",
		Version:     "1.0.0",
		Environment: os.Getenv("ENVIRONMENT"),
	})

	slog.SetDefault(log)

	log.Info("приложение запускается", "pid", os.Getpid())

	cfg := config.LoadConfig()
	log.Debug("конфигурация загружена", "server_port", cfg.ServerPort)

	db, err := postgres.NewDatabase(&cfg.Database)
	if err != nil {
		log.Error("ошибка подключения к базе данных", "error", err)
		os.Exit(1)
	}

	ctx = logger.WithLogger(ctx, log)

	log.Debug("инициализация репозиториев")
	userRepo := postgres.NewUserRepository(db)
	pvzRepo := postgres.NewPVZRepository(db)
	receptionRepo := postgres.NewReceptionRepository(db)
	productRepo := postgres.NewProductRepository(db)

	log.Debug("инициализация сервисов")
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	pvzService := services.NewPVZService(pvzRepo)
	receptionService := services.NewReceptionService(receptionRepo, pvzRepo, productRepo)
	productService := services.NewProductService(productRepo, receptionRepo, pvzRepo)

	router := api.NewRouter(authService, pvzService, receptionService, productService)

	var grpcServer *grpc.Server

	go func() {
		log.Info("gRPC сервер запускается", "port", 3000)
		grpcServer = grpc.StartGRPCServer(pvzService, 3000)
		log.Info("gRPC сервер запущен")
	}()

	router.Use(middleware.LoggingMiddleware(log))

	server := api.NewServer(cfg, router)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("HTTP сервер запускается", "port", cfg.ServerPort)
		if err := server.Start(); err != nil {
			log.Error("HTTP сервер остановлен", "error", err)
			cancel()
		}
	}()

	sig := <-quit
	log.Info("получен сигнал завершения", "signal", sig.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if grpcServer != nil {
		log.Info("завершение работы gRPC сервера...")

		done := make(chan struct{})

		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Info("gRPC сервер корректно остановлен")
		case <-shutdownCtx.Done():
			log.Warn("превышен таймаут остановки gRPC сервера, принудительное завершение")
			grpcServer.Stop()
		}
	}

	log.Info("завершение работы HTTP сервера...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("принудительное завершение сервера", "error", err)
	} else {
		log.Info("HTTP сервер корректно остановлен")
	}

	log.Info("закрытие соединения с базой данных...")
	if err := db.Close(); err != nil {
		log.Error("ошибка закрытия соединения с базой данных", "error", err)
	} else {
		log.Info("соединение с базой данных закрыто")
	}

	log.Info("приложение корректно завершило работу")
}
