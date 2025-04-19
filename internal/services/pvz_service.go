package services

import (
	"context"
	"errors"

	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"

	"github.com/google/uuid"
)

type PVZService struct {
	pvzRepo interfaces.PVZRepository
}

func NewPVZService(pvzRepo interfaces.PVZRepository) *PVZService {
	return &PVZService{
		pvzRepo: pvzRepo,
	}
}

func (s *PVZService) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	if !models.AllowedCities[city] {
		return nil, errors.New("city must be one of: Москва, Санкт-Петербург, Казань")
	}

	return s.pvzRepo.CreatePVZ(ctx, city)
}

func (s *PVZService) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	pvz, err := s.pvzRepo.GetPVZByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pvz == nil {
		return nil, errors.New("pvz not found")
	}
	return pvz, nil
}

func (s *PVZService) ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error) {
	return s.pvzRepo.ListPVZ(ctx, options)
}
