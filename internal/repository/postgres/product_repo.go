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

type ProductRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, productType models.ProductType, receptionID uuid.UUID, sequenceNum int) (*models.Product, error) {
	id := uuid.New()

	query := r.sb.Insert("products").
		Columns("id", "type", "reception_id", "sequence_num").
		Values(id, productType, receptionID, sequenceNum).
		Suffix("RETURNING id, date_time, type, reception_id, sequence_num")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var product models.Product
	err = r.db.QueryRowContext(ctx, sql, args...).Scan(
		&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var product models.Product
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting product by id: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetLastProductByReceptionID(ctx context.Context, receptionID uuid.UUID) (*models.Product, error) {
	query := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("sequence_num DESC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var product models.Product
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting last product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) DeleteProductByID(ctx context.Context, id uuid.UUID) error {
	query := r.sb.Delete("products").Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	return nil
}

func (r *ProductRepository) CountProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) (int, error) {
	query := r.sb.Select("COUNT(*)").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID})

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("error building SQL: %w", err)
	}

	var count int
	err = r.db.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting products: %w", err)
	}

	return count, nil
}

func (r *ProductRepository) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID, page, limit int) ([]*models.Product, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	query := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("sequence_num").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum); err != nil {
			return nil, 0, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, &product)
	}

	countQuery := r.sb.Select("COUNT(*)").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID})

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building count SQL: %w", err)
	}

	var total int
	err = r.db.QueryRowContext(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting products: %w", err)
	}

	return products, total, nil
}
