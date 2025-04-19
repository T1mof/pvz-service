package services

import (
	"context"
	"errors"

	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"

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
	log := logger.FromContext(ctx)
	log.Debug("AddProduct called", "pvz_id", pvzID, "product_type", productType)

	pvz, err := s.pvzRepo.GetPVZByID(ctx, pvzID)
	if err != nil {
		log.Error("Error getting PVZ", "error", err, "pvz_id", pvzID)
		return nil, err
	}
	if pvz == nil {
		log.Warn("PVZ not found", "pvz_id", pvzID)
		return nil, errors.New("pvz not found")
	}

	if productType != models.TypeElectronics && productType != models.TypeClothes && productType != models.TypeFootwear {
		log.Warn("Invalid product type", "product_type", productType)
		return nil, errors.New("invalid product type")
	}

	openReception, err := s.receptionRepo.GetLastOpenReceptionByPVZID(ctx, pvzID)
	if err != nil {
		log.Error("Error getting last open reception", "error", err, "pvz_id", pvzID)
		return nil, err
	}
	if openReception == nil {
		log.Warn("No open reception found", "pvz_id", pvzID)
		return nil, errors.New("no open reception found for this pvz")
	}

	count, err := s.productRepo.CountProductsByReceptionID(ctx, openReception.ID)
	if err != nil {
		log.Error("Error counting products", "error", err, "reception_id", openReception.ID)
		return nil, err
	}

	log.Debug("Creating product with sequence number", "reception_id", openReception.ID, "sequence_num", count+1)
	product, err := s.productRepo.CreateProduct(ctx, productType, openReception.ID, count+1)
	if err != nil {
		log.Error("Error creating product", "error", err)
		return nil, err
	}

	log.Info("Product added successfully", "product_id", product.ID, "pvz_id", pvzID, "reception_id", openReception.ID)
	return product, nil
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("DeleteLastProduct called", "pvz_id", pvzID)

	openReception, err := s.receptionRepo.GetLastOpenReceptionByPVZID(ctx, pvzID)
	if err != nil {
		log.Error("Error getting last open reception", "error", err, "pvz_id", pvzID)
		return err
	}
	if openReception == nil {
		log.Warn("No open reception found", "pvz_id", pvzID)
		return errors.New("no open reception found for this pvz")
	}

	lastProduct, err := s.productRepo.GetLastProductByReceptionID(ctx, openReception.ID)
	if err != nil {
		log.Error("Error getting last product", "error", err, "reception_id", openReception.ID)
		return err
	}
	if lastProduct == nil {
		log.Warn("No products in reception", "reception_id", openReception.ID)
		return errors.New("no products in this reception")
	}

	err = s.productRepo.DeleteProductByID(ctx, lastProduct.ID)
	if err != nil {
		log.Error("Error deleting product", "error", err, "product_id", lastProduct.ID)
		return err
	}

	log.Info("Product deleted successfully", "product_id", lastProduct.ID, "pvz_id", pvzID)
	return nil
}

func (s *ProductService) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID, page, limit int) ([]*models.Product, int, error) {
	log := logger.FromContext(ctx)
	log.Debug("GetProductsByReceptionID called", "reception_id", receptionID, "page", page, "limit", limit)

	reception, err := s.receptionRepo.GetReceptionByID(ctx, receptionID)
	if err != nil {
		log.Error("Error getting reception", "error", err, "reception_id", receptionID)
		return nil, 0, err
	}
	if reception == nil {
		log.Warn("Reception not found", "reception_id", receptionID)
		return nil, 0, errors.New("reception not found")
	}

	products, total, err := s.productRepo.GetProductsByReceptionID(ctx, receptionID, page, limit)
	if err != nil {
		log.Error("Error getting products", "error", err, "reception_id", receptionID)
		return nil, 0, err
	}

	log.Info("Products retrieved successfully", "reception_id", receptionID, "count", len(products), "total", total)
	return products, total, nil
}
