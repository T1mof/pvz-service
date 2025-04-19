package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"pvz-service/internal/domain/models"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	id := uuid.New()

	query := r.sb.Insert("users").
		Columns("id", "email", "password", "role", "created_at").
		Values(id, email, password, role, squirrel.Expr("NOW()")).
		Suffix("RETURNING id, email, role, created_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var user models.User
	err = r.db.QueryRowContext(ctx, sql, args...).Scan(
		&user.ID, &user.Email, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := r.sb.Select("id", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var user models.User
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := r.sb.Select("id", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"email": email})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var user models.User
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}

	return &user, nil
}
