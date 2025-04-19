package services

import (
	"context"
	"errors"
	"time"

	"pvz-service/internal/auth"
	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"

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
	log := logger.FromContext(ctx)
	log.Debug("Register called", "email", email, "role", role)

	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Error("Error checking existing user", "error", err)
		return nil, err
	}
	if existingUser != nil {
		log.Warn("User with this email already exists", "email", email)
		return nil, errors.New("user with this email already exists")
	}

	if role != models.RoleEmployee && role != models.RoleModerator {
		log.Warn("Invalid role provided", "role", role)
		return nil, errors.New("invalid role")
	}

	user, err := s.userRepo.CreateUser(ctx, email, password, role)
	if err != nil {
		log.Error("Error creating user", "error", err)
		return nil, err
	}

	log.Info("User registered successfully", "user_id", user.ID, "email", user.Email, "role", user.Role)
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	log := logger.FromContext(ctx)
	log.Debug("Login called", "email", email)

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Error("Error getting user by email", "error", err)
		return "", err
	}
	if user == nil {
		log.Warn("Invalid login attempt: user not found", "email", email)
		return "", errors.New("invalid email or password")
	}

	if !auth.CheckPasswordHash(password, user.Password) {
		log.Warn("Invalid login attempt: wrong password", "email", email)
		return "", errors.New("invalid email or password")
	}

	token, err := auth.GenerateToken(user, s.jwtSecret, 24*time.Hour)
	if err != nil {
		log.Error("Error generating token", "error", err)
		return "", err
	}

	log.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)
	return token, nil
}

func (s *AuthService) GenerateDummyToken(role models.UserRole) (string, error) {
	log := logger.New(logger.Config{})
	log.Debug("GenerateDummyToken called", "role", role)

	if role != models.RoleEmployee && role != models.RoleModerator {
		log.Warn("Invalid role for dummy token", "role", role)
		return "", errors.New("invalid role")
	}

	dummyUser := &models.User{
		ID:        uuid.New(),
		Email:     "dummy@example.com",
		Role:      role,
		CreatedAt: time.Now(),
	}

	token, err := auth.GenerateToken(dummyUser, s.jwtSecret, 24*time.Hour)
	if err != nil {
		log.Error("Error generating dummy token", "error", err)
		return "", err
	}

	log.Info("Dummy token generated successfully", "role", role)
	return token, nil
}

func (s *AuthService) ValidateToken(token string) (*models.User, error) {
	log := logger.New(logger.Config{})
	log.Debug("ValidateToken called")

	claims, err := auth.ValidateToken(token, s.jwtSecret)
	if err != nil {
		log.Error("Error validating token", "error", err)
		return nil, err
	}

	user := &models.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	}

	log.Info("Token validated successfully", "user_id", user.ID, "email", user.Email, "role", user.Role)
	return user, nil
}
