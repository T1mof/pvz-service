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

func setupPVZRepoTest(t *testing.T) (*PVZRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &PVZRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestCreatePVZ(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	pvzID := uuid.New()
	city := "Москва"
	regDate := time.Now()

	mock.ExpectQuery("INSERT INTO pvz").
		WithArgs(city).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, regDate, city))

	pvz, err := repo.CreatePVZ(ctx, city)

	assert.NoError(t, err)
	assert.NotNil(t, pvz)
	assert.Equal(t, pvzID, pvz.ID)
	assert.Equal(t, city, pvz.City)
	assert.WithinDuration(t, regDate, pvz.RegistrationDate, time.Second)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreatePVZ_SQLError(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	city := "Москва"

	mock.ExpectQuery("INSERT INTO pvz").
		WithArgs(city).
		WillReturnError(errors.New("database error"))

	pvz, err := repo.CreatePVZ(ctx, city)

	assert.Error(t, err)
	assert.Nil(t, pvz)
	assert.Contains(t, err.Error(), "error creating PVZ")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPVZByID(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	pvzID := uuid.New()
	city := "Санкт-Петербург"
	regDate := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, regDate, city))

	pvz, err := repo.GetPVZByID(ctx, pvzID)

	assert.NoError(t, err)
	assert.NotNil(t, pvz)
	assert.Equal(t, pvzID, pvz.ID)
	assert.Equal(t, city, pvz.City)
	assert.WithinDuration(t, regDate, pvz.RegistrationDate, time.Second)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPVZByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	pvzID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WithArgs(pvzID).
		WillReturnError(sql.ErrNoRows)

	pvz, err := repo.GetPVZByID(ctx, pvzID)

	assert.Nil(t, err)
	assert.Nil(t, pvz)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPVZByID_SQLError(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	pvzID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WithArgs(pvzID).
		WillReturnError(errors.New("database connection lost"))

	pvz, err := repo.GetPVZByID(ctx, pvzID)

	assert.Error(t, err)
	assert.Nil(t, pvz)
	assert.Contains(t, err.Error(), "error getting PVZ by id")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPVZ_NoDateFilter(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	options := models.PVZListOptions{
		Page:  1,
		Limit: 10,
	}

	pvzID := uuid.New()
	city := "Казань"
	regDate := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, regDate, city))

	receptionID := uuid.New()
	receptionDate := time.Now()
	status := "in_progress"

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, receptionDate, pvzID, status))

	productID := uuid.New()
	productType := "электроника"
	sequenceNum := 1

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(productID, time.Now(), productType, receptionID, sequenceNum))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectCommit()

	pvzs, total, err := repo.ListPVZ(ctx, options)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(pvzs))
	assert.Equal(t, 1, total)
	assert.Equal(t, pvzID, pvzs[0].PVZ.ID)
	assert.Equal(t, city, pvzs[0].PVZ.City)
	assert.Equal(t, 1, len(pvzs[0].Receptions))
	assert.Equal(t, receptionID, pvzs[0].Receptions[0].Reception.ID)
	assert.Equal(t, 1, len(pvzs[0].Receptions[0].Products))
	assert.Equal(t, productID, pvzs[0].Receptions[0].Products[0].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPVZ_WithDateFilter(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	options := models.PVZListOptions{
		Page:      1,
		Limit:     10,
		StartDate: startDate,
		EndDate:   endDate,
	}

	pvzID := uuid.New()
	city := "Москва"
	regDate := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT DISTINCT").
		WithArgs(startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, regDate, city))

	receptionID := uuid.New()
	receptionDate := time.Now()
	status := "in_progress"

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(pvzID, startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, receptionDate, pvzID, status))

	productID := uuid.New()
	productType := "электроника"

	mock.ExpectQuery("SELECT (.+) FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "sequence_num"}).
			AddRow(productID, time.Now(), productType, receptionID, 1))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(startDate, endDate).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectCommit()

	pvzs, total, err := repo.ListPVZ(ctx, options)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(pvzs))
	assert.Equal(t, 1, total)
	assert.Equal(t, pvzID, pvzs[0].PVZ.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPVZ_WithNegativePageAndLimit(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()

	options := models.PVZListOptions{
		Page:  -1,
		Limit: -5,
	}

	pvzID := uuid.New()
	city := "Казань"
	regDate := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, regDate, city))

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectCommit()

	pvzs, total, err := repo.ListPVZ(ctx, options)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(pvzs))
	assert.Equal(t, 1, total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPVZ_EmptyResult(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	options := models.PVZListOptions{
		Page:  1,
		Limit: 10,
	}

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectCommit()

	pvzs, total, err := repo.ListPVZ(ctx, options)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(pvzs))
	assert.Equal(t, 0, total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPVZ_TransactionError(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	options := models.PVZListOptions{
		Page:  1,
		Limit: 10,
	}

	mock.ExpectBegin().WillReturnError(errors.New("transaction error"))

	pvzs, total, err := repo.ListPVZ(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, pvzs)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "error starting transaction")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPVZ_QueryError(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	options := models.PVZListOptions{
		Page:  1,
		Limit: 10,
	}

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WillReturnError(errors.New("query error"))

	mock.ExpectRollback()

	pvzs, total, err := repo.ListPVZ(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, pvzs)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "error querying PVZ list")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPVZ_CountError(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	options := models.PVZListOptions{
		Page:  1,
		Limit: 10,
	}

	pvzID := uuid.New()
	city := "Москва"
	regDate := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, regDate, city))

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(errors.New("count error"))

	mock.ExpectRollback()

	pvzs, total, err := repo.ListPVZ(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, pvzs)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "error counting total PVZ")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPVZ_CommitError(t *testing.T) {
	repo, mock, cleanup := setupPVZRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	options := models.PVZListOptions{
		Page:  1,
		Limit: 10,
	}

	pvzID := uuid.New()
	city := "Москва"
	regDate := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT (.+) FROM pvz").
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(pvzID, regDate, city))

	mock.ExpectQuery("SELECT (.+) FROM receptions").
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	pvzs, total, err := repo.ListPVZ(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, pvzs)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "error committing transaction")

	assert.NoError(t, mock.ExpectationsWereMet())
}
