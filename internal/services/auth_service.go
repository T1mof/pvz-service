package services

import (
	"context"
	"errors"
	"time"

	"pvz-service/internal/auth"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo  interfaces.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo interfaces.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	if role != models.RoleEmployee && role != models.RoleModerator {
		return nil, errors.New("invalid role")
	}

	return s.userRepo.CreateUser(ctx, email, password, role)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid email or password")
	}

	if !auth.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid email or password")
	}

	token, err := auth.GenerateToken(user, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) GenerateDummyToken(role models.UserRole) (string, error) {
	if role != models.RoleEmployee && role != models.RoleModerator {
		return "", errors.New("invalid role")
	}

	dummyUser := &models.User{
		ID:        uuid.New(),
		Email:     "dummy@example.com",
		Role:      role,
		CreatedAt: time.Now(),
	}

	return auth.GenerateToken(dummyUser, s.jwtSecret, 24*time.Hour)
}

func (s *AuthService) ValidateToken(token string) (*models.User, error) {
	claims, err := auth.ValidateToken(token, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	}

	return user, nil
}
