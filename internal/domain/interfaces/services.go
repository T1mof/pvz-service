package interfaces

import (
	"context"

	"pvz-service/internal/domain/models"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, email, password string, role models.UserRole) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GenerateDummyToken(role models.UserRole) (string, error)
	ValidateToken(token string) (*models.User, error)
}

type PVZService interface {
	CreatePVZ(ctx context.Context, city string) (*models.PVZ, error)
	GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error)
	ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error)
}

type ReceptionService interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error)
}

type ProductService interface {
	AddProduct(ctx context.Context, pvzID uuid.UUID, productType models.ProductType) (*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}
