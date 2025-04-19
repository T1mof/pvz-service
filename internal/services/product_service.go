package services

import (
	"context"
	"errors"

	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"

	"github.com/google/uuid"
)

type ProductService struct {
	productRepo   interfaces.ProductRepository
	receptionRepo interfaces.ReceptionRepository
	pvzRepo       interfaces.PVZRepository
}

func NewProductService(productRepo interfaces.ProductRepository, receptionRepo interfaces.ReceptionRepository, pvzRepo interfaces.PVZRepository) *ProductService {
	return &ProductService{
		productRepo:   productRepo,
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
	}
}

func (s *ProductService) AddProduct(ctx context.Context, pvzID uuid.UUID, productType models.ProductType) (*models.Product, error) {
	pvz, err := s.pvzRepo.GetPVZByID(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if pvz == nil {
		return nil, errors.New("pvz not found")
	}

	if productType != models.TypeElectronics && productType != models.TypeClothes && productType != models.TypeFootwear {
		return nil, errors.New("invalid product type")
	}

	openReception, err := s.receptionRepo.GetLastOpenReceptionByPVZID(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if openReception == nil {
		return nil, errors.New("no open reception found for this pvz")
	}

	count, err := s.productRepo.CountProductsByReceptionID(ctx, openReception.ID)
	if err != nil {
		return nil, err
	}

	return s.productRepo.CreateProduct(ctx, productType, openReception.ID, count+1)
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	openReception, err := s.receptionRepo.GetLastOpenReceptionByPVZID(ctx, pvzID)
	if err != nil {
		return err
	}
	if openReception == nil {
		return errors.New("no open reception found for this pvz")
	}

	lastProduct, err := s.productRepo.GetLastProductByReceptionID(ctx, openReception.ID)
	if err != nil {
		return err
	}
	if lastProduct == nil {
		return errors.New("no products in this reception")
	}

	return s.productRepo.DeleteProductByID(ctx, lastProduct.ID)
}

func (s *ProductService) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID, page, limit int) ([]*models.Product, int, error) {
	reception, err := s.receptionRepo.GetReceptionByID(ctx, receptionID)
	if err != nil {
		return nil, 0, err
	}
	if reception == nil {
		return nil, 0, errors.New("reception not found")
	}

	return s.productRepo.GetProductsByReceptionID(ctx, receptionID, page, limit)
}
