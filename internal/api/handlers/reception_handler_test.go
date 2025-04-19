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

type MockReceptionService struct {
	mock.Mock
}

func (m *MockReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockReceptionService) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockReceptionService) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func setupReceptionTest() (*ReceptionHandler, *MockReceptionService) {
	mockService := new(MockReceptionService)
	handler := NewReceptionHandler(mockService)
	return handler, mockService
}

func TestCreateReception_Success(t *testing.T) {
	handler, mockService := setupReceptionTest()

	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	reception := &models.Reception{
		ID:       receptionID,
		DateTime: now,
		PVZID:    pvzID,
		Status:   models.StatusInProgress,
	}

	reqBody := models.ReceptionCreateRequest{
		PVZID: pvzID,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	mockService.On("CreateReception", mock.Anything, pvzID).Return(reception, nil)

	handler.CreateReception(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Reception
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, receptionID, response.ID)
	assert.Equal(t, pvzID, response.PVZID)
	assert.Equal(t, models.StatusInProgress, response.Status)

	mockService.AssertExpectations(t)
}

func TestCreateReception_InvalidJSON(t *testing.T) {
	handler, _ := setupReceptionTest()

	reqBody := `{"invalid json`
	req := httptest.NewRequest("POST", "/receptions", bytes.NewBufferString(reqBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	handler.CreateReception(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request format")
}

func TestCreateReception_ValidationError(t *testing.T) {
	handler, _ := setupReceptionTest()

	reqBody := models.ReceptionCreateRequest{
		PVZID: uuid.Nil,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	handler.CreateReception(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Validation failed")
}

func TestCreateReception_ServiceError(t *testing.T) {
	handler, mockService := setupReceptionTest()

	pvzID := uuid.New()

	reqBody := models.ReceptionCreateRequest{
		PVZID: pvzID,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	mockService.On("CreateReception", mock.Anything, pvzID).Return(nil, errors.New("service error"))

	handler.CreateReception(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Unable to create reception", response.Error)

	mockService.AssertExpectations(t)
}

func TestCloseLastReception_Success(t *testing.T) {
	handler, mockService := setupReceptionTest()

	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	reception := &models.Reception{
		ID:       receptionID,
		DateTime: now,
		PVZID:    pvzID,
		Status:   models.StatusClosed,
	}

	req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/close-reception", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": pvzID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("CloseLastReception", mock.Anything, pvzID).Return(reception, nil)

	handler.CloseLastReception(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Reception
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, receptionID, response.ID)
	assert.Equal(t, pvzID, response.PVZID)
	assert.Equal(t, models.StatusClosed, response.Status)

	mockService.AssertExpectations(t)
}

func TestCloseLastReception_InvalidUUID(t *testing.T) {
	handler, _ := setupReceptionTest()

	req := httptest.NewRequest("POST", "/pvz/invalid-uuid/close-reception", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": "invalid-uuid",
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	handler.CloseLastReception(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid PVZ ID format")
}

func TestCloseLastReception_ServiceError(t *testing.T) {
	handler, mockService := setupReceptionTest()

	pvzID := uuid.New()

	req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/close-reception", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": pvzID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("CloseLastReception", mock.Anything, pvzID).Return(nil, errors.New("service error"))

	handler.CloseLastReception(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Unable to close reception", response.Error)

	mockService.AssertExpectations(t)
}

func TestGetReception_Success(t *testing.T) {
	handler, mockService := setupReceptionTest()

	receptionID := uuid.New()
	pvzID := uuid.New()
	now := time.Now()

	reception := &models.Reception{
		ID:       receptionID,
		DateTime: now,
		PVZID:    pvzID,
		Status:   models.StatusInProgress,
		Products: []*models.Product{},
	}

	req := httptest.NewRequest("GET", "/receptions/"+receptionID.String(), nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"id": receptionID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("GetReceptionByID", mock.Anything, receptionID).Return(reception, nil)

	handler.GetReception(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Reception
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, receptionID, response.ID)
	assert.Equal(t, pvzID, response.PVZID)

	mockService.AssertExpectations(t)
}

func TestGetReception_InvalidUUID(t *testing.T) {
	handler, _ := setupReceptionTest()

	req := httptest.NewRequest("GET", "/receptions/invalid-uuid", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"id": "invalid-uuid",
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	handler.GetReception(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid reception ID format")
}

func TestGetReception_NotFound(t *testing.T) {
	handler, mockService := setupReceptionTest()

	receptionID := uuid.New()

	req := httptest.NewRequest("GET", "/receptions/"+receptionID.String(), nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"id": receptionID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("GetReceptionByID", mock.Anything, receptionID).Return(nil, nil)

	handler.GetReception(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Reception not found", response.Error)

	mockService.AssertExpectations(t)
}

func TestGetReception_ServiceError(t *testing.T) {
	handler, mockService := setupReceptionTest()

	receptionID := uuid.New()

	req := httptest.NewRequest("GET", "/receptions/"+receptionID.String(), nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"id": receptionID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("GetReceptionByID", mock.Anything, receptionID).Return(nil, errors.New("service error"))

	handler.GetReception(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Error retrieving reception", response.Error)

	mockService.AssertExpectations(t)
}
