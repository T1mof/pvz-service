package postgres

import (
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
)

func setupReceptionRepoTest(t *testing.T) (*ReceptionRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &ReceptionRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestCreateReception(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	pvzID := uuid.New()
	dateTime := time.Now()
	status := models.StatusInProgress

	mock.ExpectQuery("INSERT INTO receptions").
		WithArgs(pvzID, status).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, dateTime, pvzID, status))

	reception, err := repo.CreateReception(ctx, pvzID)

	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, receptionID, reception.ID)
	assert.Equal(t, pvzID, reception.PVZID)
	assert.Equal(t, status, reception.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateReception_SQLError(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	pvzID := uuid.New()

	mock.ExpectQuery("INSERT INTO receptions").
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnError(errors.New("database error"))

	reception, err := repo.CreateReception(ctx, pvzID)

	assert.Error(t, err)
	assert.Nil(t, reception)
	assert.Contains(t, err.Error(), "error creating reception")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceptionByID(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	pvzID := uuid.New()
	dateTime := time.Now()
	status := models.StatusInProgress

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, dateTime, pvzID, status))

	reception, err := repo.GetReceptionByID(ctx, receptionID)

	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, receptionID, reception.ID)
	assert.Equal(t, pvzID, reception.PVZID)
	assert.Equal(t, status, reception.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceptionByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(receptionID).
		WillReturnError(sql.ErrNoRows)

	reception, err := repo.GetReceptionByID(ctx, receptionID)

	assert.Nil(t, err)
	assert.Nil(t, reception)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastOpenReceptionByPVZID(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	pvzID := uuid.New()
	dateTime := time.Now()
	status := models.StatusInProgress

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(pvzID, status).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, dateTime, pvzID, status))

	reception, err := repo.GetLastOpenReceptionByPVZID(ctx, pvzID)

	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, receptionID, reception.ID)
	assert.Equal(t, pvzID, reception.PVZID)
	assert.Equal(t, status, reception.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastOpenReceptionByPVZID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	pvzID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnError(sql.ErrNoRows)

	reception, err := repo.GetLastOpenReceptionByPVZID(ctx, pvzID)

	assert.Nil(t, err)
	assert.Nil(t, reception)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseReception(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()

	result := sqlmock.NewResult(0, 1)

	mock.ExpectExec("UPDATE receptions").
		WithArgs(models.StatusClosed, receptionID).
		WillReturnResult(result)

	err := repo.CloseReception(ctx, receptionID)

	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseReception_SQLError(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()

	mock.ExpectExec("UPDATE receptions").
		WithArgs(models.StatusClosed, receptionID).
		WillReturnError(errors.New("database error"))

	err := repo.CloseReception(ctx, receptionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error closing reception")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListReceptions(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()

	options := ReceptionListOptions{
		Page:   1,
		Limit:  10,
		PVZID:  uuid.New(),
		Status: string(models.StatusInProgress),
	}

	receptionID := uuid.New()
	dateTime := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, dateTime, options.PVZID, options.Status))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	receptions, total, err := repo.ListReceptions(ctx, options)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(receptions))
	assert.Equal(t, 1, total)
	assert.Equal(t, receptionID, receptions[0].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListReceptions_EmptyResult(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()

	options := ReceptionListOptions{
		Page:  1,
		Limit: 10,
	}

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	receptions, total, err := repo.ListReceptions(ctx, options)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(receptions))
	assert.Equal(t, 0, total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListReceptions_QueryError(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()

	options := ReceptionListOptions{
		Page:  1,
		Limit: 10,
	}

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WillReturnError(errors.New("database error"))

	receptions, total, err := repo.ListReceptions(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, receptions)
	assert.Equal(t, 0, total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListReceptions_ScanError(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()

	options := ReceptionListOptions{
		Page:  1,
		Limit: 10,
	}

	rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
		AddRow(uuid.New(), "not-a-time-value", uuid.New(), models.StatusInProgress)

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WillReturnRows(rows)

	receptions, total, err := repo.ListReceptions(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, receptions)
	assert.Equal(t, 0, total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListReceptions_CountError(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()

	options := ReceptionListOptions{
		Page:  1,
		Limit: 10,
	}

	receptionID := uuid.New()
	dateTime := time.Now()
	pvzID := uuid.New()
	status := models.StatusInProgress

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, dateTime, pvzID, status))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(errors.New("count error"))

	receptions, total, err := repo.ListReceptions(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, receptions)
	assert.Equal(t, 0, total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceptionWithProducts(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	pvzID := uuid.New()
	dateTime := time.Now()
	status := models.StatusClosed

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, dateTime, pvzID, status))

	productID := uuid.New()
	productType := models.TypeElectronics

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(productID, time.Now(), productType, receptionID, 1))

	mock.ExpectCommit()

	reception, err := repo.GetReceptionWithProducts(ctx, receptionID)

	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, receptionID, reception.ID)
	assert.Equal(t, 1, len(reception.Products))
	assert.Equal(t, productID, reception.Products[0].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceptionWithProducts_NotFound(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(receptionID).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	reception, err := repo.GetReceptionWithProducts(ctx, receptionID)

	assert.Nil(t, err)
	assert.Nil(t, reception)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceptionWithProducts_TransactionError(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()

	mock.ExpectBegin().WillReturnError(errors.New("transaction error"))

	reception, err := repo.GetReceptionWithProducts(ctx, receptionID)

	assert.Error(t, err)
	assert.Nil(t, reception)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceptionWithProducts_CommitError(t *testing.T) {
	repo, mock, cleanup := setupReceptionRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	receptionID := uuid.New()
	pvzID := uuid.New()
	dateTime := time.Now()
	status := models.StatusClosed

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, dateTime, pvzID, status))

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}))

	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	reception, err := repo.GetReceptionWithProducts(ctx, receptionID)

	assert.Error(t, err)
	assert.Nil(t, reception)

	assert.NoError(t, mock.ExpectationsWereMet())
}
