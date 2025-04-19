package handlers

import (
	"encoding/json"
	"net/http"

	"pvz-service/internal/api/validator"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"

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
	log := logger.FromContext(r.Context())
	log.Info("запрос на добавление товара")

	var req models.ProductCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("ошибка декодирования JSON", "error", err)
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	log.Debug("запрос на добавление товара",
		"pvz_id", req.PVZID,
		"product_type", req.Type,
	)

	if err := validator.ValidateStruct(req); err != nil {
		log.Warn("ошибка валидации товара",
			"pvz_id", req.PVZID,
			"product_type", req.Type,
			"validation_errors", validator.FormatValidationErrors(err),
		)
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	product, err := h.productService.AddProduct(r.Context(), req.PVZID, req.Type)
	if err != nil {
		log.Error("ошибка добавления товара",
			"pvz_id", req.PVZID,
			"product_type", req.Type,
			"error", err,
		)
		sendErrorResponse(w, "Unable to add product", http.StatusBadRequest, err)
		return
	}

	log.Info("товар успешно добавлен",
		"product_id", product.ID,
		"pvz_id", product.ReceptionID,
		"product_type", product.Type,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	vars := mux.Vars(r)
	pvzIDStr := vars["pvzId"]

	log.Info("запрос на удаление последнего товара", "pvz_id", pvzIDStr)

	pvzID, err := uuid.Parse(pvzIDStr)
	if err != nil {
		log.Warn("некорректный формат UUID для ПВЗ", "pvz_id", pvzIDStr, "error", err)
		sendErrorResponse(w, "Invalid PVZ ID format", http.StatusBadRequest, err)
		return
	}

	err = h.productService.DeleteLastProduct(r.Context(), pvzID)
	if err != nil {
		log.Error("ошибка удаления последнего товара", "pvz_id", pvzID, "error", err)
		sendErrorResponse(w, "Unable to delete product", http.StatusBadRequest, err)
		return
	}

	log.Info("последний товар успешно удален", "pvz_id", pvzID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Product successfully deleted"})
}
