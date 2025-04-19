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

func setupUserRepoTest(t *testing.T) (*UserRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	repo := &UserRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestCreateUser(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	userID := uuid.New()
	email := "test@example.com"
	password := "hashedpassword"
	role := models.RoleEmployee
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "role", "created_at"}).
			AddRow(userID, email, role, now))

	user, err := repo.CreateUser(ctx, email, password, role)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, role, user.Role)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_SQLError(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	email := "test@example.com"
	password := "hashedpassword"
	role := models.RoleEmployee

	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(errors.New("database error"))

	user, err := repo.CreateUser(ctx, email, password, role)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "error creating user")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByID(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	userID := uuid.New()
	email := "test@example.com"
	password := "hashedpassword"
	role := models.RoleEmployee
	now := time.Now()

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"}).
			AddRow(userID, email, password, role, now))

	user, err := repo.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, password, user.Password)
	assert.Equal(t, role, user.Role)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByID(ctx, userID)

	assert.Nil(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByEmail(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	userID := uuid.New()
	email := "test@example.com"
	password := "hashedpassword"
	role := models.RoleEmployee
	now := time.Now()

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE`).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"}).
			AddRow(userID, email, password, role, now))

	user, err := repo.GetUserByEmail(ctx, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, password, user.Password)
	assert.Equal(t, role, user.Role)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	email := "test@example.com"

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByEmail(ctx, email)

	assert.Nil(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByEmail_DBError(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := createTestContext()
	email := "test@example.com"

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE`).
		WithArgs(email).
		WillReturnError(errors.New("database connection error"))

	user, err := repo.GetUserByEmail(ctx, email)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "error getting user by email")

	assert.NoError(t, mock.ExpectationsWereMet())
}
