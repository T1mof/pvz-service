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

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) AddProduct(ctx context.Context, pvzID uuid.UUID, productType models.ProductType) (*models.Product, error) {
	args := m.Called(ctx, pvzID, productType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

func (m *MockProductService) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID, page, limit int) ([]*models.Product, int, error) {
	args := m.Called(ctx, receptionID, page, limit)
	return args.Get(0).([]*models.Product), args.Int(1), args.Error(2)
}

func setupProductTest() (*ProductHandler, *MockProductService) {
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	return handler, mockService
}

func TestAddProduct_Success(t *testing.T) {
	handler, mockService := setupProductTest()

	pvzID := uuid.New()
	productType := models.TypeElectronics
	productID := uuid.New()
	productDateTime := time.Now()

	product := &models.Product{
		ID:          productID,
		DateTime:    productDateTime,
		Type:        productType,
		ReceptionID: pvzID,
		SequenceNum: 1,
	}

	reqBody := models.ProductCreateRequest{
		PVZID: pvzID,
		Type:  productType,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	mockService.On("AddProduct", mock.Anything, pvzID, productType).Return(product, nil)

	handler.AddProduct(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Product
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, productID, response.ID)
	assert.Equal(t, productType, response.Type)
	assert.Equal(t, pvzID, response.ReceptionID)

	mockService.AssertExpectations(t)
}

func TestAddProduct_InvalidJSON(t *testing.T) {
	handler, _ := setupProductTest()

	reqBody := `{"invalid json`
	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(reqBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	handler.AddProduct(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request format")
}

func TestAddProduct_ValidationError(t *testing.T) {
	handler, _ := setupProductTest()

	pvzID := uuid.New()
	reqBody := models.ProductCreateRequest{
		PVZID: pvzID,
		Type:  "invalid-type",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	handler.AddProduct(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Validation failed")
}

func TestAddProduct_ServiceError(t *testing.T) {
	handler, mockService := setupProductTest()

	pvzID := uuid.New()
	productType := models.TypeElectronics

	reqBody := models.ProductCreateRequest{
		PVZID: pvzID,
		Type:  productType,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(jsonBody))
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))
	w := httptest.NewRecorder()

	mockService.On("AddProduct", mock.Anything, pvzID, productType).Return(nil, errors.New("service error"))

	handler.AddProduct(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Unable to add product", response.Error)

	mockService.AssertExpectations(t)
}

func TestDeleteLastProduct_Success(t *testing.T) {
	handler, mockService := setupProductTest()

	pvzID := uuid.New()

	req := httptest.NewRequest("DELETE", "/products/"+pvzID.String()+"/last", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": pvzID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("DeleteLastProduct", mock.Anything, pvzID).Return(nil)

	handler.DeleteLastProduct(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Product successfully deleted", response.Message)

	mockService.AssertExpectations(t)
}

func TestDeleteLastProduct_InvalidUUID(t *testing.T) {
	handler, _ := setupProductTest()

	req := httptest.NewRequest("DELETE", "/products/invalid-uuid/last", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": "invalid-uuid",
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	handler.DeleteLastProduct(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid PVZ ID format")
}

func TestDeleteLastProduct_ServiceError(t *testing.T) {
	handler, mockService := setupProductTest()

	pvzID := uuid.New()

	req := httptest.NewRequest("DELETE", "/products/"+pvzID.String()+"/last", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), logger.New(logger.Config{Level: logger.LevelDebug, Format: "text"})))

	vars := map[string]string{
		"pvzId": pvzID.String(),
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	mockService.On("DeleteLastProduct", mock.Anything, pvzID).Return(errors.New("service error"))

	handler.DeleteLastProduct(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Unable to delete product", response.Error)

	mockService.AssertExpectations(t)
}
