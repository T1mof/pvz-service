package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvz-service/internal/domain/models"
)

var (
	productTestPvzUUID1       = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	productTestPvzUUID2       = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	productTestReceptionUUID1 = uuid.MustParse("10000000-0000-0000-0000-000000000001")
	productTestProductUUID1   = uuid.MustParse("30000000-0000-0000-0000-000000000001")
)

type ProductTestMockPVZRepository struct {
	mock.Mock
}

func (m *ProductTestMockPVZRepository) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *ProductTestMockPVZRepository) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *ProductTestMockPVZRepository) ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.PVZWithReceptionsResponse), args.Int(1), args.Error(2)
}

type ProductTestMockReceptionRepository struct {
	mock.Mock
}

func (m *ProductTestMockReceptionRepository) GetLastOpenReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *ProductTestMockReceptionRepository) CloseReception(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ProductTestMockReceptionRepository) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *ProductTestMockReceptionRepository) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *ProductTestMockReceptionRepository) GetReceptionWithProducts(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

type ProductTestMockProductRepository struct {
	mock.Mock
}

func (m *ProductTestMockProductRepository) CreateProduct(ctx context.Context, productType models.ProductType, receptionID uuid.UUID, sequenceNum int) (*models.Product, error) {
	args := m.Called(ctx, productType, receptionID, sequenceNum)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *ProductTestMockProductRepository) GetLastProductByReceptionID(ctx context.Context, receptionID uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *ProductTestMockProductRepository) DeleteProductByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ProductTestMockProductRepository) CountProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) (int, error) {
	args := m.Called(ctx, receptionID)
	return args.Int(0), args.Error(1)
}

func (m *ProductTestMockProductRepository) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *ProductTestMockProductRepository) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID, page, limit int) ([]*models.Product, int, error) {
	args := m.Called(ctx, receptionID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Product), args.Int(1), args.Error(2)
}

func setupProductTestMocks(t *testing.T) (*ProductTestMockPVZRepository, *ProductTestMockReceptionRepository, *ProductTestMockProductRepository, time.Time) {
	mockPVZRepo := new(ProductTestMockPVZRepository)
	mockReceptionRepo := new(ProductTestMockReceptionRepository)
	mockProductRepo := new(ProductTestMockProductRepository)
	now := time.Now()
	return mockPVZRepo, mockReceptionRepo, mockProductRepo, now
}

func TestProductService_AddProduct(t *testing.T) {
	testCases := []struct {
		name          string
		pvzID         uuid.UUID
		productType   models.ProductType
		setupMocks    func(*ProductTestMockPVZRepository, *ProductTestMockReceptionRepository, *ProductTestMockProductRepository, time.Time)
		expectedError bool
		checkResult   func(*testing.T, *models.Product, error)
	}{
		{
			name:        "Success - Add Electronics",
			pvzID:       productTestPvzUUID1,
			productType: models.TypeElectronics,
			setupMocks: func(pvzRepo *ProductTestMockPVZRepository, recRepo *ProductTestMockReceptionRepository, prodRepo *ProductTestMockProductRepository, now time.Time) {
				pvzRepo.On("GetPVZByID", mock.Anything, productTestPvzUUID1).Return(&models.PVZ{
					ID:               productTestPvzUUID1,
					RegistrationDate: now,
					City:             "Москва",
				}, nil)

				recRepo.On("GetLastOpenReceptionByPVZID", mock.Anything, productTestPvzUUID1).Return(&models.Reception{
					ID:       productTestReceptionUUID1,
					DateTime: now,
					PVZID:    productTestPvzUUID1,
					Status:   models.StatusInProgress,
				}, nil)

				prodRepo.On("CountProductsByReceptionID", mock.Anything, productTestReceptionUUID1).Return(5, nil)

				prodRepo.On("CreateProduct", mock.Anything, models.TypeElectronics, productTestReceptionUUID1, 6).Return(&models.Product{
					ID:          productTestProductUUID1,
					DateTime:    now,
					Type:        models.TypeElectronics,
					ReceptionID: productTestReceptionUUID1,
					SequenceNum: 6,
				}, nil)
			},
			expectedError: false,
			checkResult: func(t *testing.T, product *models.Product, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, models.TypeElectronics, product.Type)
				assert.Equal(t, productTestProductUUID1, product.ID)
			},
		},
		{
			name:        "Failure - PVZ Not Found",
			pvzID:       productTestPvzUUID2,
			productType: models.TypeElectronics,
			setupMocks: func(pvzRepo *ProductTestMockPVZRepository, recRepo *ProductTestMockReceptionRepository, prodRepo *ProductTestMockProductRepository, now time.Time) {
				pvzRepo.On("GetPVZByID", mock.Anything, productTestPvzUUID2).Return(nil, nil)
			},
			expectedError: true,
			checkResult: func(t *testing.T, product *models.Product, err error) {
				assert.Error(t, err)
				assert.Nil(t, product)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPVZRepo, mockReceptionRepo, mockProductRepo, now := setupProductTestMocks(t)
			tc.setupMocks(mockPVZRepo, mockReceptionRepo, mockProductRepo, now)

			service := NewProductService(mockProductRepo, mockReceptionRepo, mockPVZRepo)

			product, err := service.AddProduct(context.Background(), tc.pvzID, tc.productType)

			tc.checkResult(t, product, err)
			mockPVZRepo.AssertExpectations(t)
			mockReceptionRepo.AssertExpectations(t)
			mockProductRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteLastProduct(t *testing.T) {
	testCases := []struct {
		name          string
		pvzID         uuid.UUID
		setupMocks    func(*ProductTestMockPVZRepository, *ProductTestMockReceptionRepository, *ProductTestMockProductRepository, time.Time)
		expectedError bool
	}{
		{
			name:  "Success - Delete Last Product",
			pvzID: productTestPvzUUID1,
			setupMocks: func(pvzRepo *ProductTestMockPVZRepository, recRepo *ProductTestMockReceptionRepository, prodRepo *ProductTestMockProductRepository, now time.Time) {
				recRepo.On("GetLastOpenReceptionByPVZID", mock.Anything, productTestPvzUUID1).Return(&models.Reception{
					ID:       productTestReceptionUUID1,
					DateTime: now,
					PVZID:    productTestPvzUUID1,
					Status:   models.StatusInProgress,
				}, nil)

				prodRepo.On("GetLastProductByReceptionID", mock.Anything, productTestReceptionUUID1).Return(&models.Product{
					ID:          productTestProductUUID1,
					DateTime:    now,
					Type:        models.TypeElectronics,
					ReceptionID: productTestReceptionUUID1,
					SequenceNum: 5,
				}, nil)

				prodRepo.On("DeleteProductByID", mock.Anything, productTestProductUUID1).Return(nil)
			},
			expectedError: false,
		},
		{
			name:  "Failure - No Open Reception",
			pvzID: productTestPvzUUID2,
			setupMocks: func(pvzRepo *ProductTestMockPVZRepository, recRepo *ProductTestMockReceptionRepository, prodRepo *ProductTestMockProductRepository, now time.Time) {
				recRepo.On("GetLastOpenReceptionByPVZID", mock.Anything, productTestPvzUUID2).Return(nil, nil)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPVZRepo, mockReceptionRepo, mockProductRepo, now := setupProductTestMocks(t)
			tc.setupMocks(mockPVZRepo, mockReceptionRepo, mockProductRepo, now)

			service := NewProductService(mockProductRepo, mockReceptionRepo, mockPVZRepo)

			err := service.DeleteLastProduct(context.Background(), tc.pvzID)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockReceptionRepo.AssertExpectations(t)
			mockProductRepo.AssertExpectations(t)
		})
	}
}
