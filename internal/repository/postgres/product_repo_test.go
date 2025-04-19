package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"
)

func setupProductRepoTest(t *testing.T) (*ProductRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &ProductRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func createTestContext() context.Context {
	ctx := context.Background()
	testLog := logger.New(logger.Config{
		Level:  logger.LevelDebug,
		Format: "text",
		Output: nil,
	})
	return logger.WithLogger(ctx, testLog)
}

func TestCreateProduct(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	productID := uuid.New()
	now := time.Now()
	productType := models.TypeElectronics
	receptionID := uuid.New()
	sequenceNum := 1

	mock.ExpectQuery("INSERT INTO products").
		WithArgs(sqlmock.AnyArg(), productType, receptionID, sequenceNum).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(productID, now, productType, receptionID, sequenceNum))

	product, err := repo.CreateProduct(ctx, productType, receptionID, sequenceNum)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, productType, product.Type)
	assert.Equal(t, receptionID, product.ReceptionID)
	assert.Equal(t, sequenceNum, product.SequenceNum)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProduct_Error(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	productType := models.TypeElectronics
	receptionID := uuid.New()
	sequenceNum := 1

	mock.ExpectQuery("INSERT INTO products").
		WithArgs(sqlmock.AnyArg(), productType, receptionID, sequenceNum).
		WillReturnError(errors.New("database error"))

	product, err := repo.CreateProduct(ctx, productType, receptionID, sequenceNum)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "error creating product")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductByID(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	productID := uuid.New()
	now := time.Now()
	productType := models.TypeElectronics
	receptionID := uuid.New()
	sequenceNum := 1

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(productID, now, productType, receptionID, sequenceNum))

	product, err := repo.GetProductByID(ctx, productID)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, productID, product.ID)
	assert.Equal(t, productType, product.Type)
	assert.Equal(t, receptionID, product.ReceptionID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	productID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(productID).
		WillReturnError(sql.ErrNoRows)

	product, err := repo.GetProductByID(ctx, productID)

	assert.Nil(t, err)
	assert.Nil(t, product)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastProductByReceptionID(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	productID := uuid.New()
	now := time.Now()
	productType := models.TypeElectronics
	receptionID := uuid.New()
	sequenceNum := 5

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(productID, now, productType, receptionID, sequenceNum))

	product, err := repo.GetLastProductByReceptionID(ctx, receptionID)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, productID, product.ID)
	assert.Equal(t, productType, product.Type)
	assert.Equal(t, sequenceNum, product.SequenceNum)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastProductByReceptionID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnError(sql.ErrNoRows)

	product, err := repo.GetLastProductByReceptionID(ctx, receptionID)

	assert.Nil(t, err)
	assert.Nil(t, product)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProductByID(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	productID := uuid.New()

	result := sqlmock.NewResult(0, 1)

	mock.ExpectExec("DELETE FROM products").
		WithArgs(productID).
		WillReturnResult(result)

	err := repo.DeleteProductByID(ctx, productID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProductByID_Error(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	productID := uuid.New()

	mock.ExpectExec("DELETE FROM products").
		WithArgs(productID).
		WillReturnError(errors.New("database error"))

	err := repo.DeleteProductByID(ctx, productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error deleting product")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCountProductsByReceptionID(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	expectedCount := 10

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(expectedCount))

	count, err := repo.CountProductsByReceptionID(ctx, receptionID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCountProductsByReceptionID_Error(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(receptionID).
		WillReturnError(errors.New("database error"))

	count, err := repo.CountProductsByReceptionID(ctx, receptionID)

	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "error counting products")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductsByReceptionID(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	page := 1
	limit := 10

	product1ID := uuid.New()
	product2ID := uuid.New()
	now := time.Now()
	productType := models.TypeElectronics
	total := 2

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(product1ID, now, productType, receptionID, 1).
			AddRow(product2ID, now, productType, receptionID, 2))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(total))

	products, totalCount, err := repo.GetProductsByReceptionID(ctx, receptionID, page, limit)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, total, totalCount)

	if len(products) >= 2 {
		assert.Equal(t, product1ID, products[0].ID)
		assert.Equal(t, product2ID, products[1].ID)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductsByReceptionID_NegativePageAndLimit(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	page := -1
	limit := -5

	productID := uuid.New()
	now := time.Now()
	productType := models.TypeElectronics

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(productID, now, productType, receptionID, 1))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(1))

	products, totalCount, err := repo.GetProductsByReceptionID(ctx, receptionID, page, limit)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(products))
	assert.Equal(t, 1, totalCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductsByReceptionID_QueryError(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	page := 1
	limit := 10

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnError(errors.New("database error"))

	products, totalCount, err := repo.GetProductsByReceptionID(ctx, receptionID, page, limit)

	assert.Error(t, err)
	assert.Nil(t, products)
	assert.Equal(t, 0, totalCount)
	assert.Contains(t, err.Error(), "error querying products")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductsByReceptionID_ScanError(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	page := 1
	limit := 10

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time"}).
			AddRow(uuid.New(), time.Now()))

	products, totalCount, err := repo.GetProductsByReceptionID(ctx, receptionID, page, limit)

	assert.Error(t, err)
	assert.Nil(t, products)
	assert.Equal(t, 0, totalCount)
	assert.Contains(t, err.Error(), "error scanning product row")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductsByReceptionID_CountError(t *testing.T) {
	repo, mock, cleanup := setupProductRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	page := 1
	limit := 10

	productID := uuid.New()
	now := time.Now()
	productType := models.TypeElectronics

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(productID, now, productType, receptionID, 1))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(receptionID).
		WillReturnError(errors.New("count error"))

	products, totalCount, err := repo.GetProductsByReceptionID(ctx, receptionID, page, limit)

	assert.Error(t, err)
	assert.Nil(t, products)
	assert.Equal(t, 0, totalCount)
	assert.Contains(t, err.Error(), "error counting products")

	assert.NoError(t, mock.ExpectationsWereMet())
}
