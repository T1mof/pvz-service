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

type ReceptionHandler struct {
	receptionService interfaces.ReceptionService
}

func NewReceptionHandler(receptionService interfaces.ReceptionService) *ReceptionHandler {
	return &ReceptionHandler{
		receptionService: receptionService,
	}
}

func (h *ReceptionHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	var req models.ReceptionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	reception, err := h.receptionService.CreateReception(r.Context(), req.PVZID)
	if err != nil {
		sendErrorResponse(w, "Unable to create reception", http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reception)
}

func (h *ReceptionHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pvzIDStr := vars["pvzId"]

	pvzID, err := uuid.Parse(pvzIDStr)
	if err != nil {
		sendErrorResponse(w, "Invalid PVZ ID format", http.StatusBadRequest, err)
		return
	}

	reception, err := h.receptionService.CloseLastReception(r.Context(), pvzID)
	if err != nil {
		sendErrorResponse(w, "Unable to close reception", http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reception)
}

func (h *ReceptionHandler) GetReception(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendErrorResponse(w, "Invalid reception ID format", http.StatusBadRequest, err)
		return
	}

	reception, err := h.receptionService.GetReceptionByID(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, "Error retrieving reception", http.StatusInternalServerError, err)
		return
	}

	if reception == nil {
		sendErrorResponse(w, "Reception not found", http.StatusNotFound, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reception)
}
