package services

import (
	"context"
	"errors"

	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"
	"pvz-service/internal/metrics"

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
	log := logger.FromContext(ctx)
	log.Debug("CreatePVZ called", "city", city)

	if !models.AllowedCities[city] {
		log.Warn("Invalid city provided", "city", city)
		return nil, errors.New("city must be one of: Москва, Санкт-Петербург, Казань")
	}

	pvz, err := s.pvzRepo.CreatePVZ(ctx, city)
	if err != nil {
		log.Error("Error creating PVZ", "error", err, "city", city)
		return nil, err
	}

	metrics.IncrementPVZCreated()

	log.Info("PVZ created successfully", "pvz_id", pvz.ID, "city", pvz.City)
	return pvz, nil
}

func (s *PVZService) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	log := logger.FromContext(ctx)
	log.Debug("GetPVZByID called", "pvz_id", id)

	pvz, err := s.pvzRepo.GetPVZByID(ctx, id)
	if err != nil {
		log.Error("Error getting PVZ", "error", err, "pvz_id", id)
		return nil, err
	}
	if pvz == nil {
		log.Warn("PVZ not found", "pvz_id", id)
		return nil, errors.New("pvz not found")
	}

	log.Info("PVZ retrieved successfully", "pvz_id", pvz.ID, "city", pvz.City)
	return pvz, nil
}

func (s *PVZService) ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error) {
	log := logger.FromContext(ctx)
	log.Debug("ListPVZ called",
		"page", options.Page,
		"limit", options.Limit,
		"has_start_date", !options.StartDate.IsZero(),
		"has_end_date", !options.EndDate.IsZero(),
	)

	pvzs, total, err := s.pvzRepo.ListPVZ(ctx, options)
	if err != nil {
		log.Error("Error listing PVZs", "error", err)
		return nil, 0, err
	}

	log.Info("PVZs listed successfully", "count", len(pvzs), "total", total)
	return pvzs, total, nil
}
