package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	args := m.Called(ctx, email, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateDummyToken(role models.UserRole) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func setupTest() (*AuthHandler, *MockAuthService) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	return handler, mockService
}

func setupTestContext() {
	logConfig := logger.Config{
		Level:  logger.LevelDebug,
		Format: "text",
		Output: nil,
	}
	_ = logger.New(logConfig)
}

func TestRegister_Success(t *testing.T) {
	setupTestContext()
	handler, mockService := setupTest()

	userID := uuid.New()
	userEmail := "test@example.com"
	userPassword := "password123"
	userRole := models.RoleEmployee

	user := &models.User{
		ID:    userID,
		Email: userEmail,
		Role:  userRole,
	}

	reqBody := models.AuthRequest{
		Email:    userEmail,
		Password: userPassword,
		Role:     userRole,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mockService.On("Register", mock.Anything, userEmail, userPassword, userRole).Return(user, nil)

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, userID, response.ID)
	assert.Equal(t, userEmail, response.Email)
	assert.Equal(t, userRole, response.Role)

	mockService.AssertExpectations(t)
}

func TestRegister_InvalidJSON(t *testing.T) {
	setupTestContext()
	handler, _ := setupTest()

	reqBody := `{"invalid json`
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request format")
}

func TestRegister_ValidationError(t *testing.T) {
	setupTestContext()
	handler, _ := setupTest()

	reqBody := models.AuthRequest{
		Email:    "invalid-email",
		Password: "",
		Role:     "",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Validation failed")
}

func TestRegister_ServiceError(t *testing.T) {
	setupTestContext()
	handler, mockService := setupTest()

	userEmail := "test@example.com"
	userPassword := "password123"
	userRole := models.RoleEmployee

	reqBody := models.AuthRequest{
		Email:    userEmail,
		Password: userPassword,
		Role:     userRole,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mockService.On("Register", mock.Anything, userEmail, userPassword, userRole).
		Return(nil, errors.New("user already exists"))

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Registration failed", response.Error)

	mockService.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	setupTestContext()
	handler, mockService := setupTest()

	userEmail := "test@example.com"
	userPassword := "password123"
	token := "jwt.token.string"

	reqBody := models.AuthRequest{
		Email:    userEmail,
		Password: userPassword,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mockService.On("Login", mock.Anything, userEmail, userPassword).Return(token, nil)

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, token, response.Token)

	mockService.AssertExpectations(t)
}

func TestLogin_InvalidJSON(t *testing.T) {
	setupTestContext()
	handler, _ := setupTest()

	reqBody := `{"invalid json`
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request format")
}

func TestLogin_ValidationError(t *testing.T) {
	setupTestContext()
	handler, _ := setupTest()

	reqBody := models.AuthRequest{
		Email:    "invalid-email",
		Password: "",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Validation failed")
}

func TestLogin_ServiceError(t *testing.T) {
	setupTestContext()
	handler, mockService := setupTest()

	userEmail := "test@example.com"
	userPassword := "password123"

	reqBody := models.AuthRequest{
		Email:    userEmail,
		Password: userPassword,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mockService.On("Login", mock.Anything, userEmail, userPassword).
		Return("", errors.New("invalid credentials"))

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response.Error)

	mockService.AssertExpectations(t)
}

func TestDummyLogin_Success(t *testing.T) {
	setupTestContext()
	handler, mockService := setupTest()

	role := models.RoleEmployee
	token := "jwt.dummy.token"

	reqBody := struct {
		Role string `json:"role"`
	}{
		Role: string(role),
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/dummy-login", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mockService.On("GenerateDummyToken", role).Return(token, nil)

	handler.DummyLogin(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, token, response.Token)

	mockService.AssertExpectations(t)
}

func TestDummyLogin_InvalidJSON(t *testing.T) {
	setupTestContext()
	handler, _ := setupTest()

	reqBody := `{"invalid json`
	req := httptest.NewRequest("POST", "/auth/dummy-login", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handler.DummyLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request format")
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	setupTestContext()
	handler, _ := setupTest()

	reqBody := struct {
		Role string `json:"role"`
	}{
		Role: "invalid-role",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/dummy-login", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.DummyLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid role")
}

func TestDummyLogin_ServiceError(t *testing.T) {
	setupTestContext()
	handler, mockService := setupTest()

	role := models.RoleEmployee

	reqBody := struct {
		Role string `json:"role"`
	}{
		Role: string(role),
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/dummy-login", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mockService.On("GenerateDummyToken", role).
		Return("", errors.New("token generation failed"))

	handler.DummyLogin(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Failed to generate token", response.Error)

	mockService.AssertExpectations(t)
}
