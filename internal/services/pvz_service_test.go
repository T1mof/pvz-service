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
	pvzTestUUID1           = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	pvzTestNonexistentUUID = uuid.MustParse("99999999-9999-9999-9999-999999999999")
)

type PVZTestMockRepository struct {
	mock.Mock
}

func (m *PVZTestMockRepository) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *PVZTestMockRepository) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *PVZTestMockRepository) ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]*models.PVZWithReceptionsResponse), args.Int(1), args.Error(2)
}

func TestPVZService_CreatePVZ(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name          string
		city          string
		mockSetup     func(*PVZTestMockRepository)
		expectedPVZ   *models.PVZ
		expectedError bool
	}{
		{
			name: "Success - Moscow",
			city: "Москва",
			mockSetup: func(repo *PVZTestMockRepository) {
				repo.On("CreatePVZ", mock.Anything, "Москва").
					Return(&models.PVZ{
						ID:               pvzTestUUID1,
						RegistrationDate: now,
						City:             "Москва",
					}, nil)
			},
			expectedPVZ: &models.PVZ{
				ID:               pvzTestUUID1,
				RegistrationDate: now,
				City:             "Москва",
			},
			expectedError: false,
		},
		{
			name: "Failure - Invalid City",
			city: "Новосибирск",
			mockSetup: func(repo *PVZTestMockRepository) {
			},
			expectedPVZ:   nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(PVZTestMockRepository)
			tc.mockSetup(mockRepo)
			service := NewPVZService(mockRepo)

			pvz, err := service.CreatePVZ(context.Background(), tc.city)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, pvz)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pvz)
				assert.Equal(t, tc.expectedPVZ.ID, pvz.ID)
				assert.Equal(t, tc.expectedPVZ.City, pvz.City)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_GetPVZByID(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name          string
		pvzID         uuid.UUID
		mockSetup     func(*PVZTestMockRepository)
		expectedPVZ   *models.PVZ
		expectedError bool
	}{
		{
			name:  "Success - PVZ Found",
			pvzID: pvzTestUUID1,
			mockSetup: func(repo *PVZTestMockRepository) {
				repo.On("GetPVZByID", mock.Anything, pvzTestUUID1).
					Return(&models.PVZ{
						ID:               pvzTestUUID1,
						RegistrationDate: now,
						City:             "Москва",
					}, nil)
			},
			expectedPVZ: &models.PVZ{
				ID:               pvzTestUUID1,
				RegistrationDate: now,
				City:             "Москва",
			},
			expectedError: false,
		},
		{
			name:  "Failure - PVZ Not Found",
			pvzID: pvzTestNonexistentUUID,
			mockSetup: func(repo *PVZTestMockRepository) {
				repo.On("GetPVZByID", mock.Anything, pvzTestNonexistentUUID).
					Return(nil, nil)
			},
			expectedPVZ:   nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(PVZTestMockRepository)
			tc.mockSetup(mockRepo)
			service := NewPVZService(mockRepo)

			pvz, err := service.GetPVZByID(context.Background(), tc.pvzID)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, pvz)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pvz)
				assert.Equal(t, tc.expectedPVZ.ID, pvz.ID)
				assert.Equal(t, tc.expectedPVZ.City, pvz.City)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_ListPVZ(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name          string
		options       models.PVZListOptions
		mockSetup     func(*PVZTestMockRepository)
		expectedTotal int
		expectedError bool
	}{
		{
			name: "Success - List All PVZs",
			options: models.PVZListOptions{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func(repo *PVZTestMockRepository) {
				pvzs := []*models.PVZWithReceptionsResponse{
					{
						PVZ: &models.PVZ{
							ID:               pvzTestUUID1,
							RegistrationDate: now,
							City:             "Москва",
						},
						Receptions: []*models.ReceptionWithProducts{},
					},
				}
				repo.On("ListPVZ", mock.Anything, mock.Anything).Return(pvzs, 1, nil)
			},
			expectedTotal: 1,
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(PVZTestMockRepository)
			tc.mockSetup(mockRepo)
			service := NewPVZService(mockRepo)

			pvzs, total, err := service.ListPVZ(context.Background(), tc.options)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, pvzs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pvzs)
				assert.Equal(t, tc.expectedTotal, total)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
