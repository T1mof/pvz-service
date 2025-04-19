package services

import (
	"context"
	"errors"

	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"

	"github.com/google/uuid"
)

type ReceptionService struct {
	receptionRepo interfaces.ReceptionRepository
	pvzRepo       interfaces.PVZRepository
	productRepo   interfaces.ProductRepository
}

func NewReceptionService(receptionRepo interfaces.ReceptionRepository, pvzRepo interfaces.PVZRepository, productRepo interfaces.ProductRepository) *ReceptionService {
	return &ReceptionService{
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
		productRepo:   productRepo,
	}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	log := logger.FromContext(ctx)
	log.Debug("CreateReception called", "pvz_id", pvzID)

	pvz, err := s.pvzRepo.GetPVZByID(ctx, pvzID)
	if err != nil {
		log.Error("Error getting PVZ", "error", err, "pvz_id", pvzID)
		return nil, err
	}
	if pvz == nil {
		log.Warn("PVZ not found", "pvz_id", pvzID)
		return nil, errors.New("pvz not found")
	}

	openReception, err := s.receptionRepo.GetLastOpenReceptionByPVZID(ctx, pvzID)
	if err != nil {
		log.Error("Error checking for open receptions", "error", err, "pvz_id", pvzID)
		return nil, err
	}
	if openReception != nil {
		log.Warn("Open reception already exists", "pvz_id", pvzID, "reception_id", openReception.ID)
		return nil, errors.New("there is already an open reception for this pvz")
	}

	reception, err := s.receptionRepo.CreateReception(ctx, pvzID)
	if err != nil {
		log.Error("Error creating reception", "error", err, "pvz_id", pvzID)
		return nil, err
	}

	log.Info("Reception created successfully", "reception_id", reception.ID, "pvz_id", pvzID)
	return reception, nil
}

func (s *ReceptionService) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	log := logger.FromContext(ctx)
	log.Debug("CloseLastReception called", "pvz_id", pvzID)

	openReception, err := s.receptionRepo.GetLastOpenReceptionByPVZID(ctx, pvzID)
	if err != nil {
		log.Error("Error getting last open reception", "error", err, "pvz_id", pvzID)
		return nil, err
	}
	if openReception == nil {
		log.Warn("No open reception found", "pvz_id", pvzID)
		return nil, errors.New("no open reception found for this pvz")
	}

	err = s.receptionRepo.CloseReception(ctx, openReception.ID)
	if err != nil {
		log.Error("Error closing reception", "error", err, "reception_id", openReception.ID)
		return nil, err
	}

	updatedReception, err := s.receptionRepo.GetReceptionByID(ctx, openReception.ID)
	if err != nil {
		log.Error("Error getting updated reception", "error", err, "reception_id", openReception.ID)
		return nil, err
	}

	log.Info("Reception closed successfully", "reception_id", updatedReception.ID, "pvz_id", pvzID)
	return updatedReception, nil
}

func (s *ReceptionService) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	log := logger.FromContext(ctx)
	log.Debug("GetReceptionByID called", "reception_id", id)

	reception, err := s.receptionRepo.GetReceptionByID(ctx, id)
	if err != nil {
		log.Error("Error getting reception", "error", err, "reception_id", id)
		return nil, err
	}
	if reception == nil {
		log.Warn("Reception not found", "reception_id", id)
		return nil, errors.New("reception not found")
	}

	products, _, err := s.productRepo.GetProductsByReceptionID(ctx, id, 1, 1000)
	if err != nil {
		log.Error("Error getting products for reception", "error", err, "reception_id", id)
		return nil, err
	}

	reception.Products = products
	log.Info("Reception retrieved successfully", "reception_id", id, "products_count", len(products))
	return reception, nil
}
