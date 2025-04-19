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

type ReceptionHandler struct {
	receptionService interfaces.ReceptionService
}

func NewReceptionHandler(receptionService interfaces.ReceptionService) *ReceptionHandler {
	return &ReceptionHandler{
		receptionService: receptionService,
	}
}

func (h *ReceptionHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("запрос на создание приемки")

	var req models.ReceptionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("ошибка декодирования JSON", "error", err)
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	log.Debug("запрос на создание приемки", "pvz_id", req.PVZID)

	if err := validator.ValidateStruct(req); err != nil {
		log.Warn("ошибка валидации приемки",
			"pvz_id", req.PVZID,
			"validation_errors", validator.FormatValidationErrors(err),
		)
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	reception, err := h.receptionService.CreateReception(r.Context(), req.PVZID)
	if err != nil {
		log.Error("ошибка создания приемки", "pvz_id", req.PVZID, "error", err)
		sendErrorResponse(w, "Unable to create reception", http.StatusBadRequest, err)
		return
	}

	log.Info("приемка успешно создана",
		"reception_id", reception.ID,
		"pvz_id", reception.PVZID,
		"status", reception.Status,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reception)
}

func (h *ReceptionHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	vars := mux.Vars(r)
	pvzIDStr := vars["pvzId"]

	log.Info("запрос на закрытие последней приемки", "pvz_id", pvzIDStr)

	pvzID, err := uuid.Parse(pvzIDStr)
	if err != nil {
		log.Warn("некорректный формат UUID для ПВЗ", "pvz_id", pvzIDStr, "error", err)
		sendErrorResponse(w, "Invalid PVZ ID format", http.StatusBadRequest, err)
		return
	}

	reception, err := h.receptionService.CloseLastReception(r.Context(), pvzID)
	if err != nil {
		log.Error("ошибка закрытия последней приемки", "pvz_id", pvzID, "error", err)
		sendErrorResponse(w, "Unable to close reception", http.StatusBadRequest, err)
		return
	}

	log.Info("последняя приемка успешно закрыта",
		"reception_id", reception.ID,
		"pvz_id", reception.PVZID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reception)
}

func (h *ReceptionHandler) GetReception(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	vars := mux.Vars(r)
	idStr := vars["id"]

	log.Info("запрос на получение приемки", "reception_id", idStr)

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn("некорректный формат UUID для приемки", "reception_id", idStr, "error", err)
		sendErrorResponse(w, "Invalid reception ID format", http.StatusBadRequest, err)
		return
	}

	reception, err := h.receptionService.GetReceptionByID(r.Context(), id)
	if err != nil {
		log.Error("ошибка получения приемки", "reception_id", id, "error", err)
		sendErrorResponse(w, "Error retrieving reception", http.StatusInternalServerError, err)
		return
	}

	if reception == nil {
		log.Warn("приемка не найдена", "reception_id", id)
		sendErrorResponse(w, "Reception not found", http.StatusNotFound, nil)
		return
	}

	log.Info("приемка успешно получена",
		"reception_id", id,
		"pvz_id", reception.PVZID,
		"status", reception.Status,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reception)
}
