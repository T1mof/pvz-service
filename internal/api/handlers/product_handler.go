package handlers

import (
	"encoding/json"
	"net/http"

	"pvz-service/internal/api/validator"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productService interfaces.ProductService
}

// SuccessResponse для стандартизации успешных ответов
type SuccessResponse struct {
	Message string `json:"message"`
}

func NewProductHandler(productService interfaces.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var req models.ProductCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	// Валидация входящих данных
	if err := validator.ValidateStruct(req); err != nil {
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	// Добавление товара
	product, err := h.productService.AddProduct(r.Context(), req.PVZID, req.Type)
	if err != nil {
		// Используем общее сообщение для клиента
		sendErrorResponse(w, "Unable to add product", http.StatusBadRequest, err)
		return
	}

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pvzIDStr := vars["pvzId"]

	// Преобразование строки в UUID
	pvzID, err := uuid.Parse(pvzIDStr)
	if err != nil {
		sendErrorResponse(w, "Invalid PVZ ID format", http.StatusBadRequest, err)
		return
	}

	// Удаление последнего товара
	err = h.productService.DeleteLastProduct(r.Context(), pvzID)
	if err != nil {
		sendErrorResponse(w, "Unable to delete product", http.StatusBadRequest, err)
		return
	}

	// Отправка стандартизированного ответа об успехе
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Product successfully deleted"})
}
