package services

import (
	"context"
	"errors"

	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"

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
	pvz, err := s.pvzRepo.GetPVZByID(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if pvz == nil {
		return nil, errors.New("pvz not found")
	}

	openReception, err := s.receptionRepo.GetLastOpenReceptionByPVZID(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if openReception != nil {
		return nil, errors.New("there is already an open reception for this pvz")
	}

	return s.receptionRepo.CreateReception(ctx, pvzID)
}

func (s *ReceptionService) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	openReception, err := s.receptionRepo.GetLastOpenReceptionByPVZID(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if openReception == nil {
		return nil, errors.New("no open reception found for this pvz")
	}

	err = s.receptionRepo.CloseReception(ctx, openReception.ID)
	if err != nil {
		return nil, err
	}

	return s.receptionRepo.GetReceptionByID(ctx, openReception.ID)
}

func (s *ReceptionService) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	reception, err := s.receptionRepo.GetReceptionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if reception == nil {
		return nil, errors.New("reception not found")
	}

	products, _, err := s.productRepo.GetProductsByReceptionID(ctx, id, 1, 1000)
	if err != nil {
		return nil, err
	}

	reception.Products = products
	return reception, nil
}
