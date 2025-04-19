package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvz-service/internal/domain/models"
)

var (
	pvzServiceTestUUID1           = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	pvzServiceTestNonexistentUUID = uuid.MustParse("99999999-9999-9999-9999-999999999999")
)

type PVZServiceTestMockRepository struct {
	mock.Mock
}

func (m *PVZServiceTestMockRepository) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *PVZServiceTestMockRepository) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *PVZServiceTestMockRepository) ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]*models.PVZWithReceptionsResponse), args.Int(1), args.Error(2)
}

func setupPVZServiceTest(t *testing.T) (*PVZServiceTestMockRepository, *PVZService, time.Time) {
	mockRepo := new(PVZServiceTestMockRepository)
	service := NewPVZService(mockRepo)
	now := time.Now()
	return mockRepo, service, now
}

func TestPVZServiceCreate(t *testing.T) {
	testCases := []struct {
		name          string
		city          string
		setupMock     func(*PVZServiceTestMockRepository, time.Time)
		expectedError bool
		checkResult   func(*testing.T, *models.PVZ, error)
	}{
		{
			name: "Success - Moscow",
			city: "Москва",
			setupMock: func(repo *PVZServiceTestMockRepository, now time.Time) {
				repo.On("CreatePVZ", mock.Anything, "Москва").
					Return(&models.PVZ{
						ID:               pvzServiceTestUUID1,
						RegistrationDate: now,
						City:             "Москва",
					}, nil)
			},
			expectedError: false,
			checkResult: func(t *testing.T, pvz *models.PVZ, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, pvz)
				assert.Equal(t, "Москва", pvz.City)
			},
		},
		{
			name: "Failure - Invalid City",
			city: "Новосибирск",
			setupMock: func(repo *PVZServiceTestMockRepository, now time.Time) {
			},
			expectedError: true,
			checkResult: func(t *testing.T, pvz *models.PVZ, err error) {
				assert.Error(t, err)
				assert.Nil(t, pvz)
				assert.Contains(t, err.Error(), "city must be one of")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, service, now := setupPVZServiceTest(t)
			tc.setupMock(repo, now)

			pvz, err := service.CreatePVZ(context.Background(), tc.city)

			tc.checkResult(t, pvz, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestPVZServiceGetByID(t *testing.T) {
	testCases := []struct {
		name          string
		pvzID         uuid.UUID
		setupMock     func(*PVZServiceTestMockRepository, time.Time)
		expectedError bool
		checkResult   func(*testing.T, *models.PVZ, error)
	}{
		{
			name:  "Success - PVZ Found",
			pvzID: pvzServiceTestUUID1,
			setupMock: func(repo *PVZServiceTestMockRepository, now time.Time) {
				repo.On("GetPVZByID", mock.Anything, pvzServiceTestUUID1).
					Return(&models.PVZ{
						ID:               pvzServiceTestUUID1,
						RegistrationDate: now,
						City:             "Москва",
					}, nil)
			},
			expectedError: false,
			checkResult: func(t *testing.T, pvz *models.PVZ, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, pvz)
				assert.Equal(t, pvzServiceTestUUID1, pvz.ID)
			},
		},
		{
			name:  "Failure - PVZ Not Found",
			pvzID: pvzServiceTestNonexistentUUID,
			setupMock: func(repo *PVZServiceTestMockRepository, now time.Time) {
				repo.On("GetPVZByID", mock.Anything, pvzServiceTestNonexistentUUID).
					Return(nil, nil)
			},
			expectedError: true,
			checkResult: func(t *testing.T, pvz *models.PVZ, err error) {
				assert.Error(t, err)
				assert.Nil(t, pvz)
				assert.Contains(t, err.Error(), "not found")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, service, now := setupPVZServiceTest(t)
			tc.setupMock(repo, now)

			pvz, err := service.GetPVZByID(context.Background(), tc.pvzID)

			tc.checkResult(t, pvz, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestPVZServiceList(t *testing.T) {
	testCases := []struct {
		name          string
		options       models.PVZListOptions
		setupMock     func(*PVZServiceTestMockRepository, time.Time)
		expectedError bool
		checkResult   func(*testing.T, []*models.PVZWithReceptionsResponse, int, error)
	}{
		{
			name: "Success - List PVZs",
			options: models.PVZListOptions{
				Page:  1,
				Limit: 10,
			},
			setupMock: func(repo *PVZServiceTestMockRepository, now time.Time) {
				pvzs := []*models.PVZWithReceptionsResponse{
					{
						PVZ: &models.PVZ{
							ID:               pvzServiceTestUUID1,
							RegistrationDate: now,
							City:             "Москва",
						},
						Receptions: []*models.ReceptionWithProducts{},
					},
				}
				repo.On("ListPVZ", mock.Anything, mock.Anything).Return(pvzs, 1, nil)
			},
			expectedError: false,
			checkResult: func(t *testing.T, pvzs []*models.PVZWithReceptionsResponse, total int, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, pvzs)
				assert.Equal(t, 1, total)
				assert.Len(t, pvzs, 1)
				assert.Equal(t, "Москва", pvzs[0].PVZ.City)
			},
		},
		{
			name: "Failure - Database Error",
			options: models.PVZListOptions{
				Page:  1,
				Limit: 10,
			},
			setupMock: func(repo *PVZServiceTestMockRepository, now time.Time) {
				repo.On("ListPVZ", mock.Anything, mock.Anything).
					Return(([]*models.PVZWithReceptionsResponse)(nil), 0, errors.New("database error"))
			},
			expectedError: true,
			checkResult: func(t *testing.T, pvzs []*models.PVZWithReceptionsResponse, total int, err error) {
				assert.Error(t, err)
				assert.Nil(t, pvzs)
				assert.Equal(t, 0, total)
				assert.Contains(t, err.Error(), "database error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, service, now := setupPVZServiceTest(t)
			tc.setupMock(repo, now)

			pvzs, total, err := service.ListPVZ(context.Background(), tc.options)

			tc.checkResult(t, pvzs, total, err)
			repo.AssertExpectations(t)
		})
	}
}
