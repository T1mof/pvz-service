package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pvz-service/internal/api"
	"pvz-service/internal/config"
	"pvz-service/internal/repository/postgres"
	"pvz-service/internal/services"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Подключаемся к базе данных
	db, err := postgres.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Закрытие DB перенесено из defer, будет выполняться явно при завершении

	// Инициализируем репозитории
	userRepo := postgres.NewUserRepository(db)
	pvzRepo := postgres.NewPVZRepository(db)
	receptionRepo := postgres.NewReceptionRepository(db)
	productRepo := postgres.NewProductRepository(db)

	// Инициализируем сервисы
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	pvzService := services.NewPVZService(pvzRepo)
	receptionService := services.NewReceptionService(receptionRepo, pvzRepo, productRepo)
	productService := services.NewProductService(productRepo, receptionRepo, pvzRepo)

	// Инициализируем маршрутизатор и сервер
	router := api.NewRouter(authService, pvzService, receptionService, productService)
	server := api.NewServer(cfg, router)

	// Создаем канал для перехвата сигналов
	quit := make(chan os.Signal, 1)
	// Подписываемся на сигналы завершения: Ctrl+C, systemd stop, docker stop и т.д.
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Server starting on port %d", cfg.ServerPort)
		if err := server.Start(); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// Ожидаем сигнал от ОС
	<-quit
	log.Println("Shutting down server...")

	// Создаем контекст с таймаутом для завершения
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Корректно завершаем работу сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Закрываем соединение с базой данных
	log.Println("Closing database connection...")
	if err := db.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	log.Println("Server exited gracefully")
}
