package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvz-service/internal/auth"
	"pvz-service/internal/domain/models"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	args := m.Called(ctx, email, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	userUUID1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userUUID2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	userUUID3 := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	testCases := []struct {
		name          string
		email         string
		password      string
		role          models.UserRole
		mockSetup     func(*MockUserRepository)
		expectedUser  *models.User
		expectedError bool
	}{
		{
			name:     "Success - New Employee",
			email:    "employee@example.com",
			password: "password123",
			role:     models.RoleEmployee,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "employee@example.com").Return(nil, nil)
				repo.On("CreateUser", mock.Anything, "employee@example.com", "password123", models.RoleEmployee).
					Return(&models.User{
						ID:        userUUID1,
						Email:     "employee@example.com",
						Role:      models.RoleEmployee,
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedUser: &models.User{
				ID:    userUUID1,
				Email: "employee@example.com",
				Role:  models.RoleEmployee,
			},
			expectedError: false,
		},
		{
			name:     "Success - New Moderator",
			email:    "moderator@example.com",
			password: "password123",
			role:     models.RoleModerator,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "moderator@example.com").Return(nil, nil)
				repo.On("CreateUser", mock.Anything, "moderator@example.com", "password123", models.RoleModerator).
					Return(&models.User{
						ID:        userUUID2,
						Email:     "moderator@example.com",
						Role:      models.RoleModerator,
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedUser: &models.User{
				ID:    userUUID2,
				Email: "moderator@example.com",
				Role:  models.RoleModerator,
			},
			expectedError: false,
		},
		{
			name:     "Failure - User Already Exists",
			email:    "existing@example.com",
			password: "password123",
			role:     models.RoleEmployee,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "existing@example.com").
					Return(&models.User{
						ID:        userUUID3,
						Email:     "existing@example.com",
						Role:      models.RoleEmployee,
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedUser:  nil,
			expectedError: true,
		},
		{
			name:     "Failure - Invalid Role",
			email:    "test@example.com",
			password: "password123",
			role:     "invalid_role",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(nil, nil)
			},
			expectedUser:  nil,
			expectedError: true,
		},
		{
			name:     "Failure - Database Error",
			email:    "error@example.com",
			password: "password123",
			role:     models.RoleEmployee,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "error@example.com").Return(nil, nil)
				repo.On("CreateUser", mock.Anything, "error@example.com", "password123", models.RoleEmployee).
					Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			service := NewAuthService(mockRepo, "test_jwt_secret")

			user, err := service.Register(context.Background(), tc.email, tc.password, tc.role)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tc.expectedUser.Email, user.Email)
				assert.Equal(t, tc.expectedUser.Role, user.Role)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	hashedPassword, _ := auth.HashPassword("password123")

	userUUID1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	testCases := []struct {
		name          string
		email         string
		password      string
		mockSetup     func(*MockUserRepository)
		expectedToken bool
		expectedError bool
	}{
		{
			name:     "Success - Valid Credentials",
			email:    "user@example.com",
			password: "password123",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "user@example.com").
					Return(&models.User{
						ID:        userUUID1,
						Email:     "user@example.com",
						Password:  hashedPassword,
						Role:      models.RoleEmployee,
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedToken: true,
			expectedError: false,
		},
		{
			name:     "Failure - User Not Found",
			email:    "nonexistent@example.com",
			password: "password123",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)
			},
			expectedToken: false,
			expectedError: true,
		},
		{
			name:     "Failure - Invalid Password",
			email:    "user@example.com",
			password: "wrongpassword",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "user@example.com").
					Return(&models.User{
						ID:        userUUID1,
						Email:     "user@example.com",
						Password:  hashedPassword,
						Role:      models.RoleEmployee,
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedToken: false,
			expectedError: true,
		},
		{
			name:     "Failure - Database Error",
			email:    "error@example.com",
			password: "password123",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", mock.Anything, "error@example.com").
					Return(nil, errors.New("database error"))
			},
			expectedToken: false,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			service := NewAuthService(mockRepo, "test_jwt_secret")
			token, err := service.Login(context.Background(), tc.email, tc.password)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_GenerateDummyToken(t *testing.T) {
	testCases := []struct {
		name          string
		role          models.UserRole
		expectedError bool
	}{
		{
			name:          "Success - Employee Token",
			role:          models.RoleEmployee,
			expectedError: false,
		},
		{
			name:          "Success - Moderator Token",
			role:          models.RoleModerator,
			expectedError: false,
		},
		{
			name:          "Failure - Invalid Role",
			role:          "invalid_role",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			service := NewAuthService(mockRepo, "test_jwt_secret")

			token, err := service.GenerateDummyToken(tc.role)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				user, validateErr := service.ValidateToken(token)
				assert.NoError(t, validateErr)
				assert.Equal(t, tc.role, user.Role)
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test_jwt_secret")

	validToken, _ := service.GenerateDummyToken(models.RoleEmployee)

	testCases := []struct {
		name          string
		token         string
		expectedUser  *models.User
		expectedError bool
	}{
		{
			name:  "Success - Valid Token",
			token: validToken,
			expectedUser: &models.User{
				Role: models.RoleEmployee,
			},
			expectedError: false,
		},
		{
			name:          "Failure - Empty Token",
			token:         "",
			expectedUser:  nil,
			expectedError: true,
		},
		{
			name:          "Failure - Invalid Token",
			token:         "invalid.token.string",
			expectedUser:  nil,
			expectedError: true,
		},
		{
			name:          "Failure - Malformed Token",
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedUser:  nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := service.ValidateToken(tc.token)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tc.expectedUser.Role, user.Role)
			}
		})
	}
}
