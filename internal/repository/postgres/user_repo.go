package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"

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
	log := logger.FromContext(ctx)
	log.Debug("создание пользователя",
		"email", email,
		"role", role,
	)

	id := uuid.New()

	query := r.sb.Insert("users").
		Columns("id", "email", "password", "role", "created_at").
		Values(id, email, password, role, squirrel.Expr("NOW()")).
		Suffix("RETURNING id, email, role, created_at")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var user models.User
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&user.ID, &user.Email, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		log.Error("ошибка создания пользователя в БД",
			"error", err,
			"email", email,
		)
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	log.Info("пользователь успешно создан",
		"user_id", user.ID,
		"email", user.Email,
		"role", user.Role,
	)

	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение пользователя по ID", "user_id", id)

	query := r.sb.Select("id", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "user_id", id)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var user models.User
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("пользователь не найден", "user_id", id)
			return nil, nil
		}
		log.Error("ошибка получения пользователя", "error", err, "user_id", id)
		return nil, fmt.Errorf("error getting user by id: %w", err)
	}

	log.Debug("пользователь успешно получен",
		"user_id", user.ID,
		"email", user.Email,
		"role", user.Role,
	)

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение пользователя по email", "email", email)

	query := r.sb.Select("id", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"email": email})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "email", email)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var user models.User
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("пользователь не найден по email", "email", email)
			return nil, nil
		}
		log.Error("ошибка получения пользователя по email", "error", err, "email", email)
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}

	log.Debug("пользователь успешно получен по email",
		"user_id", user.ID,
		"email", user.Email,
		"role", user.Role,
	)

	return &user, nil
}
