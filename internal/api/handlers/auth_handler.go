package handlers

import (
	"encoding/json"
	"net/http"

	"pvz-service/internal/api/validator"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"

	"golang.org/x/exp/slog"
)

type AuthHandler struct {
	authService interfaces.AuthService
}

// Структура для стандартизированных ответов об ошибках
type ErrorResponse struct {
	Error string `json:"error"`
}

func sendErrorResponse(w http.ResponseWriter, message string, status int, err error) {
	// Используем глобальный логгер, так как у нас нет доступа к контексту запроса
	log := slog.Default()

	if err != nil {
		log.Error("ошибка обработки запроса",
			"error", err,
			"status", status,
			"message", message,
		)
	} else {
		log.Warn("запрос завершен с ошибкой",
			"status", status,
			"message", message,
		)
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
	log := logger.FromContext(r.Context())
	log.Info("запрос на регистрацию пользователя")

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("ошибка декодирования JSON", "error", err)
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	log.Debug("запрос на регистрацию", "email", req.Email, "role", req.Role)

	if err := validator.ValidateStruct(req); err != nil {
		log.Warn("ошибка валидации при регистрации",
			"email", req.Email,
			"validation_errors", validator.FormatValidationErrors(err),
		)
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		log.Error("ошибка регистрации пользователя",
			"email", req.Email,
			"role", req.Role,
			"error", err,
		)
		sendErrorResponse(w, "Registration failed", http.StatusBadRequest, err)
		return
	}

	log.Info("пользователь успешно зарегистрирован",
		"user_id", user.ID,
		"email", user.Email,
		"role", user.Role,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("запрос на аутентификацию")

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("ошибка декодирования JSON", "error", err)
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	// Для безопасности логируем только email
	log.Debug("попытка входа", "email", req.Email)

	if err := validator.ValidateStruct(req); err != nil {
		log.Warn("ошибка валидации при входе",
			"email", req.Email,
			"validation_errors", validator.FormatValidationErrors(err),
		)
		sendErrorResponse(w, "Validation failed: "+validator.FormatValidationErrors(err), http.StatusBadRequest, nil)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		// Для защиты от атак перечисления пользователей не логируем причину ошибки
		log.Warn("неудачная попытка входа", "email", req.Email)
		sendErrorResponse(w, "Invalid credentials", http.StatusUnauthorized, err)
		return
	}

	log.Info("пользователь успешно аутентифицирован", "email", req.Email)

	tokenResponse := models.TokenResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse)
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("запрос на тестовую аутентификацию")

	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("ошибка декодирования JSON", "error", err)
		sendErrorResponse(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	log.Debug("запрос тестового токена", "requested_role", req.Role)

	var role models.UserRole
	if req.Role == string(models.RoleModerator) {
		role = models.RoleModerator
	} else if req.Role == string(models.RoleEmployee) {
		role = models.RoleEmployee
	} else {
		log.Warn("запрошена недопустимая роль", "role", req.Role)
		sendErrorResponse(w, "Invalid role: must be 'employee' or 'moderator'", http.StatusBadRequest, nil)
		return
	}

	token, err := h.authService.GenerateDummyToken(role)
	if err != nil {
		log.Error("ошибка генерации тестового токена", "role", role, "error", err)
		sendErrorResponse(w, "Failed to generate token", http.StatusInternalServerError, err)
		return
	}

	log.Info("тестовый токен успешно сгенерирован", "role", role)

	tokenResponse := models.TokenResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse)
}
