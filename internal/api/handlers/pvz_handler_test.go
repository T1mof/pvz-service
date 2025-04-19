package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"
)

type MockPVZService struct {
	mock.Mock
}

func (m *MockPVZService) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *MockPVZService) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *MockPVZService) ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.PVZWithReceptionsResponse), args.Int(1), args.Error(2)
}

func setupPVZTest() (*PVZHandler, *MockPVZService) {
	mockService := new(MockPVZService)
	handler := NewPVZHandler(mockService)
	return handler, mockService
}

func TestCreatePVZ_Success(t *testing.T) {
	handler, mockService := setupPVZTest()

	pvzID := uuid.New()
	city := "Москва"
	registrationDate := time.Now()

	pvz := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: registrationDate,
		City:             city,
	}

	reqBody := models.PVZCreateRequest{
		City: city,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	mockService.On("CreatePVZ", mock.Anything, city).Return(pvz, nil)

	handler.CreatePVZ(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.PVZ
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, pvzID, response.ID)
	assert.Equal(t, city, response.City)

	mockService.AssertExpectations(t)
}

func TestCreatePVZ_InvalidJSON(t *testing.T) {
	handler, _ := setupPVZTest()

	reqBody := `{"invalid json`
	req := httptest.NewRequest("POST", "/pvz", bytes.NewBufferString(reqBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	handler.CreatePVZ(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request format")
}

func TestCreatePVZ_ValidationError(t *testing.T) {
	handler, _ := setupPVZTest()

	reqBody := models.PVZCreateRequest{
		City: "",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	handler.CreatePVZ(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Validation failed")
}

func TestCreatePVZ_ServiceError(t *testing.T) {
	handler, mockService := setupPVZTest()

	city := "Москва"

	reqBody := models.PVZCreateRequest{
		City: city,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	mockService.On("CreatePVZ", mock.Anything, city).Return(nil, errors.New("service error"))

	handler.CreatePVZ(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Unable to create PVZ", response.Error)

	mockService.AssertExpectations(t)
}

func TestListPVZ_Success(t *testing.T) {
	handler, mockService := setupPVZTest()

	pvzID := uuid.New()
	city := "Москва"
	registrationDate := time.Now()

	pvz := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: registrationDate,
		City:             city,
	}

	pvzs := []*models.PVZWithReceptionsResponse{
		{
			PVZ:        pvz,
			Receptions: []*models.ReceptionWithProducts{},
		},
	}

	total := 1
	page := 1
	limit := 10

	options := models.PVZListOptions{
		Page:  page,
		Limit: limit,
	}

	req := httptest.NewRequest("GET", "/pvz?page=1&limit=10", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	mockService.On("ListPVZ", mock.Anything, options).Return(pvzs, total, nil)

	handler.ListPVZ(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["pagination"])

	pagination := response["pagination"].(map[string]interface{})
	assert.Equal(t, float64(page), pagination["page"])
	assert.Equal(t, float64(limit), pagination["limit"])
	assert.Equal(t, float64(total), pagination["total"])

	mockService.AssertExpectations(t)
}

func TestListPVZ_InvalidDateFormat(t *testing.T) {
	handler, _ := setupPVZTest()

	req := httptest.NewRequest("GET", "/pvz?startDate=invalid-date", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	handler.ListPVZ(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid startDate format")
}

func TestListPVZ_ServiceError(t *testing.T) {
	handler, mockService := setupPVZTest()

	options := models.PVZListOptions{
		Page:  1,
		Limit: 10,
	}

	req := httptest.NewRequest("GET", "/pvz?page=1&limit=10", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	// Использование пустого слайса вместо nil
	mockService.On("ListPVZ", mock.Anything, options).Return([]*models.PVZWithReceptionsResponse{}, 0, errors.New("service error"))

	handler.ListPVZ(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Failed to retrieve PVZ list", response.Error)

	mockService.AssertExpectations(t)
}

func TestGetPVZByID_Success(t *testing.T) {
	handler, mockService := setupPVZTest()

	pvzID := uuid.New()
	city := "Москва"
	registrationDate := time.Now()

	pvz := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: registrationDate,
		City:             city,
	}

	req := httptest.NewRequest("GET", "/pvz/"+pvzID.String(), nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": pvzID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

	handler.GetPVZByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PVZ
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, pvzID, response.ID)
	assert.Equal(t, city, response.City)

	mockService.AssertExpectations(t)
}

func TestGetPVZByID_InvalidUUID(t *testing.T) {
	handler, _ := setupPVZTest()

	req := httptest.NewRequest("GET", "/pvz/invalid-uuid", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": "invalid-uuid",
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	handler.GetPVZByID(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid PVZ ID format")
}

func TestGetPVZByID_NotFound(t *testing.T) {
	handler, mockService := setupPVZTest()

	pvzID := uuid.New()

	req := httptest.NewRequest("GET", "/pvz/"+pvzID.String(), nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": pvzID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("GetPVZByID", mock.Anything, pvzID).Return(nil, nil)

	handler.GetPVZByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "PVZ not found", response.Error)

	mockService.AssertExpectations(t)
}

func TestGetPVZByID_ServiceError(t *testing.T) {
	handler, mockService := setupPVZTest()

	pvzID := uuid.New()

	req := httptest.NewRequest("GET", "/pvz/"+pvzID.String(), nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": pvzID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("GetPVZByID", mock.Anything, pvzID).Return(nil, errors.New("service error"))

	handler.GetPVZByID(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Error retrieving PVZ", response.Error)

	mockService.AssertExpectations(t)
}
