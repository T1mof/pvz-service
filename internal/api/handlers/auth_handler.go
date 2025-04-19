package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"pvz-service/internal/api/validator"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
)

type AuthHandler struct {
	authService interfaces.AuthService
}

// Структура для стандартизированных ответов об ошибках
type ErrorResponse struct {
	Error string `json:"error"`
}

// Вспомогательная функция для отправки JSON-ответа с ошибкой
func sendErrorResponse(w http.ResponseWriter, message string, status int, err error) {
	if err != nil {
		log.Printf("Error details: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func NewAuthHandler(authService interfaces.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		sendErrorResponse(w, "Registration failed", http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		sendErrorResponse(w, "Invalid credentials", http.StatusUnauthorized, err)
		return
	}

	tokenResponse := models.TokenResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse)
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	var role models.UserRole
	if req.Role == string(models.RoleModerator) {
		role = models.RoleModerator
	} else if req.Role == string(models.RoleEmployee) {
		role = models.RoleEmployee
	} else {
		sendErrorResponse(w, "Invalid role: must be 'employee' or 'moderator'", http.StatusBadRequest, nil)
		return
	}

	token, err := h.authService.GenerateDummyToken(role)
	if err != nil {
		sendErrorResponse(w, "Failed to generate token", http.StatusInternalServerError, err)
		return
	}

	tokenResponse := models.TokenResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse)
}
