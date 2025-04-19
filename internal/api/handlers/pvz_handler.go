package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"pvz-service/internal/api/validator"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PVZHandler struct {
	pvzService interfaces.PVZService
}

func NewPVZHandler(pvzService interfaces.PVZService) *PVZHandler {
	return &PVZHandler{
		pvzService: pvzService,
	}
}

func (h *PVZHandler) CreatePVZ(w http.ResponseWriter, r *http.Request) {
	var req models.PVZCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	pvz, err := h.pvzService.CreatePVZ(r.Context(), req.City)
	if err != nil {
		sendErrorResponse(w, "Unable to create PVZ", http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pvz)
}

func (h *PVZHandler) ListPVZ(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 30 {
			limit = l
		}
	}

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			sendErrorResponse(w, "Invalid startDate format. Use RFC3339 format", http.StatusBadRequest, err)
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			sendErrorResponse(w, "Invalid endDate format. Use RFC3339 format", http.StatusBadRequest, err)
			return
		}
	}

	options := models.PVZListOptions{
		Page:      page,
		Limit:     limit,
		StartDate: startDate,
		EndDate:   endDate,
	}

	pvzs, total, err := h.pvzService.ListPVZ(r.Context(), options)
	if err != nil {
		sendErrorResponse(w, "Failed to retrieve PVZ list", http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"data": pvzs,
		"pagination": map[string]int{
			"page":      page,
			"limit":     limit,
			"total":     total,
			"pageCount": (total + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PVZHandler) GetPVZByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["pvzId"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendErrorResponse(w, "Invalid PVZ ID format", http.StatusBadRequest, err)
		return
	}

	pvz, err := h.pvzService.GetPVZByID(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, "Error retrieving PVZ", http.StatusInternalServerError, err)
		return
	}

	if pvz == nil {
		sendErrorResponse(w, "PVZ not found", http.StatusNotFound, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pvz)
}
