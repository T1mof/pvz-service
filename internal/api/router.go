package api

import (
	"net/http"

	"pvz-service/internal/api/handlers"
	"pvz-service/internal/api/middleware"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"

	"github.com/gorilla/mux"
)

func NewRouter(
	authService interfaces.AuthService,
	pvzService interfaces.PVZService,
	receptionService interfaces.ReceptionService,
	productService interfaces.ProductService,
) *mux.Router {
	router := mux.NewRouter()

	// Добавляем общий middleware для мониторинга производительности
	router.Use(middleware.ResponseTimeMiddleware)
	router.Use(middleware.RecoveryMiddleware)

	// Инициализируем обработчики
	authHandler := handlers.NewAuthHandler(authService)
	pvzHandler := handlers.NewPVZHandler(pvzService)
	receptionHandler := handlers.NewReceptionHandler(receptionService)
	productHandler := handlers.NewProductHandler(productService)

	// Создаем middleware для авторизации
	authMiddleware := middleware.AuthMiddleware(authService)
	employeeRoleMiddleware := middleware.RequireRole(models.RoleEmployee)
	moderatorRoleMiddleware := middleware.RequireRole(models.RoleModerator)

	// Авторизация - согласно спецификации
	router.HandleFunc("/dummyLogin", authHandler.DummyLogin).Methods("POST")
	router.HandleFunc("/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/login", authHandler.Login).Methods("POST")

	// ПВЗ - согласно спецификации
	pvzRouter := router.PathPrefix("/pvz").Subrouter()
	pvzRouter.Use(authMiddleware)

	// POST /pvz - создание ПВЗ (только модератор)
	pvzRouter.Handle("", moderatorRoleMiddleware(http.HandlerFunc(pvzHandler.CreatePVZ))).Methods("POST")

	// GET /pvz - получение списка ПВЗ
	pvzRouter.HandleFunc("", pvzHandler.ListPVZ).Methods("GET")

	// POST /pvz/{pvzId}/close_last_reception - закрытие последней приемки (employee)
	router.Handle("/pvz/{pvzId}/close_last_reception",
		authMiddleware(employeeRoleMiddleware(http.HandlerFunc(receptionHandler.CloseLastReception)))).Methods("POST")

	// POST /pvz/{pvzId}/delete_last_product - удаление последнего товара (employee)
	router.Handle("/pvz/{pvzId}/delete_last_product",
		authMiddleware(employeeRoleMiddleware(http.HandlerFunc(productHandler.DeleteLastProduct)))).Methods("POST")

	// POST /receptions - создание новой приемки (employee)
	router.Handle("/receptions",
		authMiddleware(employeeRoleMiddleware(http.HandlerFunc(receptionHandler.CreateReception)))).Methods("POST")

	// POST /products - добавление товара (employee)
	router.Handle("/products",
		authMiddleware(employeeRoleMiddleware(http.HandlerFunc(productHandler.AddProduct)))).Methods("POST")

	return router
}
