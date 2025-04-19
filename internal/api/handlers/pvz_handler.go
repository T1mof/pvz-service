package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"pvz-service/internal/api/validator"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"

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
	log := logger.FromContext(r.Context())
	log.Info("запрос на создание ПВЗ")

	var req models.PVZCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("ошибка декодирования JSON", "error", err)
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	log.Debug("запрос на создание ПВЗ", "city", req.City)

	if err := validator.ValidateStruct(req); err != nil {
		log.Warn("ошибка валидации ПВЗ",
			"city", req.City,
			"validation_errors", validator.FormatValidationErrors(err),
		)
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	pvz, err := h.pvzService.CreatePVZ(r.Context(), req.City)
	if err != nil {
		log.Error("ошибка создания ПВЗ", "city", req.City, "error", err)
		sendErrorResponse(w, "Unable to create PVZ", http.StatusBadRequest, err)
		return
	}

	log.Info("ПВЗ успешно создан", "pvz_id", pvz.ID, "city", pvz.City)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pvz)
}

func (h *PVZHandler) ListPVZ(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	log.Info("запрос на получение списка ПВЗ",
		"page", pageStr,
		"limit", limitStr,
		"startDate", startDateStr,
		"endDate", endDateStr,
	)

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		} else if err != nil {
			log.Warn("некорректное значение page", "page", pageStr, "error", err)
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 30 {
			limit = l
		} else if err != nil {
			log.Warn("некорректное значение limit", "limit", limitStr, "error", err)
		}
	}

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			log.Warn("некорректный формат startDate", "startDate", startDateStr, "error", err)
			sendErrorResponse(w, "Invalid startDate format. Use RFC3339 format", http.StatusBadRequest, err)
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			log.Warn("некорректный формат endDate", "endDate", endDateStr, "error", err)
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

	log.Debug("получение списка ПВЗ с параметрами",
		"page", page,
		"limit", limit,
		"startDate", startDate,
		"endDate", endDate,
	)

	pvzs, total, err := h.pvzService.ListPVZ(r.Context(), options)
	if err != nil {
		log.Error("ошибка получения списка ПВЗ", "error", err)
		sendErrorResponse(w, "Failed to retrieve PVZ list", http.StatusInternalServerError, err)
		return
	}

	log.Info("список ПВЗ успешно получен",
		"count", len(pvzs),
		"total", total,
	)

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
	log := logger.FromContext(r.Context())

	vars := mux.Vars(r)
	idStr := vars["pvzId"]

	log.Info("запрос на получение ПВЗ по ID", "pvz_id", idStr)

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn("некорректный формат UUID", "pvz_id", idStr, "error", err)
		sendErrorResponse(w, "Invalid PVZ ID format", http.StatusBadRequest, err)
		return
	}

	pvz, err := h.pvzService.GetPVZByID(r.Context(), id)
	if err != nil {
		log.Error("ошибка получения ПВЗ", "pvz_id", id, "error", err)
		sendErrorResponse(w, "Error retrieving PVZ", http.StatusInternalServerError, err)
		return
	}

	if pvz == nil {
		log.Warn("ПВЗ не найден", "pvz_id", id)
		sendErrorResponse(w, "PVZ not found", http.StatusNotFound, nil)
		return
	}

	log.Info("ПВЗ успешно получен", "pvz_id", id, "city", pvz.City)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pvz)
}
