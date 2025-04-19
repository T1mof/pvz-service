package interfaces

import (
	"context"

	"pvz-service/internal/domain/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type PVZRepository interface {
	CreatePVZ(ctx context.Context, city string) (*models.PVZ, error)
	GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error)
	ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error)
}

type ReceptionRepository interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error)
	GetLastOpenReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, id uuid.UUID) error
	GetReceptionWithProducts(ctx context.Context, id uuid.UUID) (*models.Reception, error)
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, productType models.ProductType, receptionID uuid.UUID, sequenceNum int) (*models.Product, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetLastProductByReceptionID(ctx context.Context, receptionID uuid.UUID) (*models.Product, error)
	DeleteProductByID(ctx context.Context, id uuid.UUID) error
	CountProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) (int, error)
	GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID, page, limit int) ([]*models.Product, int, error)
}
