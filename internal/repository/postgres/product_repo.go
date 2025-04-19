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
	log := logger.FromContext(ctx)
	log.Debug("создание товара",
		"product_type", productType,
		"reception_id", receptionID,
		"sequence_num", sequenceNum,
	)

	id := uuid.New()

	query := r.sb.Insert("products").
		Columns("id", "type", "reception_id", "sequence_num").
		Values(id, productType, receptionID, sequenceNum).
		Suffix("RETURNING id, date_time, type, reception_id, sequence_num")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	if log.Enabled(ctx, logger.LevelDebug) {
		log.Debug("SQL запрос", "query", sqlQuery)
	}

	var product models.Product
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum,
	)

	if err != nil {
		log.Error("ошибка создания товара в БД",
			"error", err,
			"product_type", productType,
			"reception_id", receptionID,
		)
		return nil, fmt.Errorf("error creating product: %w", err)
	}

	log.Info("товар успешно создан",
		"product_id", product.ID,
		"product_type", product.Type,
		"reception_id", product.ReceptionID,
	)

	return &product, nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение товара по ID", "product_id", id)

	query := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "product_id", id)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var product models.Product
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("товар не найден", "product_id", id)
			return nil, nil
		}
		log.Error("ошибка получения товара", "error", err, "product_id", id)
		return nil, fmt.Errorf("error getting product by id: %w", err)
	}

	log.Debug("товар успешно получен",
		"product_id", product.ID,
		"product_type", product.Type,
		"reception_id", product.ReceptionID,
	)

	return &product, nil
}

func (r *ProductRepository) GetLastProductByReceptionID(ctx context.Context, receptionID uuid.UUID) (*models.Product, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение последнего товара для приемки", "reception_id", receptionID)

	query := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("sequence_num DESC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "reception_id", receptionID)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var product models.Product
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("товары для приемки не найдены", "reception_id", receptionID)
			return nil, nil
		}
		log.Error("ошибка получения последнего товара", "error", err, "reception_id", receptionID)
		return nil, fmt.Errorf("error getting last product: %w", err)
	}

	log.Debug("последний товар успешно получен",
		"product_id", product.ID,
		"product_type", product.Type,
		"sequence_num", product.SequenceNum,
	)

	return &product, nil
}

func (r *ProductRepository) DeleteProductByID(ctx context.Context, id uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("удаление товара", "product_id", id)

	query := r.sb.Delete("products").Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "product_id", id)
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка удаления товара", "error", err, "product_id", id)
		return fmt.Errorf("error deleting product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Warn("не удалось получить количество затронутых строк", "error", err)
	} else if rowsAffected == 0 {
		log.Warn("товар не найден при удалении", "product_id", id)
	} else {
		log.Info("товар успешно удален", "product_id", id, "rows_affected", rowsAffected)
	}

	return nil
}

func (r *ProductRepository) CountProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) (int, error) {
	log := logger.FromContext(ctx)
	log.Debug("подсчет товаров для приемки", "reception_id", receptionID)

	query := r.sb.Select("COUNT(*)").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "reception_id", receptionID)
		return 0, fmt.Errorf("error building SQL: %w", err)
	}

	var count int
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&count)
	if err != nil {
		log.Error("ошибка подсчета товаров", "error", err, "reception_id", receptionID)
		return 0, fmt.Errorf("error counting products: %w", err)
	}

	log.Debug("подсчет товаров завершен", "reception_id", receptionID, "count", count)
	return count, nil
}

func (r *ProductRepository) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID, page, limit int) ([]*models.Product, int, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение списка товаров для приемки",
		"reception_id", receptionID,
		"page", page,
		"limit", limit,
	)

	if limit <= 0 {
		limit = 10
		log.Debug("установлено значение limit по умолчанию", "limit", limit)
	}
	if page <= 0 {
		page = 1
		log.Debug("установлено значение page по умолчанию", "page", page)
	}

	offset := (page - 1) * limit

	query := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("sequence_num").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "reception_id", receptionID)
		return nil, 0, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка выполнения запроса товаров", "error", err, "reception_id", receptionID)
		return nil, 0, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum); err != nil {
			log.Error("ошибка сканирования строки товара", "error", err)
			return nil, 0, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, &product)
	}

	countQuery := r.sb.Select("COUNT(*)").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID})

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL для подсчета", "error", err, "reception_id", receptionID)
		return nil, 0, fmt.Errorf("error building count SQL: %w", err)
	}

	var total int
	err = r.db.QueryRowContext(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		log.Error("ошибка подсчета товаров", "error", err, "reception_id", receptionID)
		return nil, 0, fmt.Errorf("error counting products: %w", err)
	}

	log.Info("список товаров успешно получен",
		"reception_id", receptionID,
		"count", len(products),
		"total", total,
	)

	return products, total, nil
}
