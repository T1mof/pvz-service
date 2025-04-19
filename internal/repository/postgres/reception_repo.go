package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pvz-service/internal/domain/models"
	"pvz-service/internal/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type ReceptionRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

func NewReceptionRepository(db *sql.DB) *ReceptionRepository {
	return &ReceptionRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ReceptionRepository) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	log := logger.FromContext(ctx)
	log.Debug("создание приемки", "pvz_id", pvzID)

	query := r.sb.Insert("receptions").
		Columns("pvz_id", "status").
		Values(pvzID, models.StatusInProgress).
		Suffix("RETURNING id, date_time, pvz_id, status")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	if log.Enabled(ctx, logger.LevelDebug) {
		log.Debug("SQL запрос", "query", sqlQuery, "args", args)
	}

	var reception models.Reception
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status,
	)

	if err != nil {
		log.Error("ошибка создания приемки в БД", "error", err, "pvz_id", pvzID)
		return nil, fmt.Errorf("error creating reception: %w", err)
	}

	log.Info("приемка успешно создана",
		"reception_id", reception.ID,
		"pvz_id", reception.PVZID,
		"status", reception.Status,
	)

	return &reception, nil
}

func (r *ReceptionRepository) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение приемки по ID", "reception_id", id)

	query := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "reception_id", id)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var reception models.Reception
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("приемка не найдена", "reception_id", id)
			return nil, nil
		}
		log.Error("ошибка получения приемки", "error", err, "reception_id", id)
		return nil, fmt.Errorf("error getting reception by id: %w", err)
	}

	log.Debug("приемка успешно получена",
		"reception_id", reception.ID,
		"pvz_id", reception.PVZID,
		"status", reception.Status,
	)

	return &reception, nil
}

func (r *ReceptionRepository) GetLastOpenReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение последней открытой приемки для ПВЗ", "pvz_id", pvzID)

	query := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.And{
			squirrel.Eq{"pvz_id": pvzID},
			squirrel.Eq{"status": models.StatusInProgress},
		}).
		OrderBy("date_time DESC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "pvz_id", pvzID)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var reception models.Reception
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("открытая приемка не найдена для ПВЗ", "pvz_id", pvzID)
			return nil, nil
		}
		log.Error("ошибка получения последней открытой приемки", "error", err, "pvz_id", pvzID)
		return nil, fmt.Errorf("error getting last open reception: %w", err)
	}

	log.Debug("последняя открытая приемка успешно получена",
		"reception_id", reception.ID,
		"pvz_id", reception.PVZID,
	)

	return &reception, nil
}

func (r *ReceptionRepository) CloseReception(ctx context.Context, id uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("закрытие приемки", "reception_id", id)

	query := r.sb.Update("receptions").
		Set("status", models.StatusClosed).
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "reception_id", id)
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка закрытия приемки", "error", err, "reception_id", id)
		return fmt.Errorf("error closing reception: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Warn("не удалось получить количество затронутых строк", "error", err)
	} else if rowsAffected == 0 {
		log.Warn("приемка не найдена при закрытии", "reception_id", id)
	} else {
		log.Info("приемка успешно закрыта", "reception_id", id)
	}

	return nil
}

type ReceptionListOptions struct {
	Page     int
	Limit    int
	PVZID    uuid.UUID
	Status   string
	FromDate time.Time
	ToDate   time.Time
}

func (r *ReceptionRepository) ListReceptions(ctx context.Context, options ReceptionListOptions) ([]*models.Reception, int, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение списка приемок",
		"page", options.Page,
		"limit", options.Limit,
		"pvz_id", options.PVZID,
		"status", options.Status,
		"has_from_date", !options.FromDate.IsZero(),
		"has_to_date", !options.ToDate.IsZero(),
	)

	if options.Limit <= 0 {
		options.Limit = 10
		log.Debug("установлено значение limit по умолчанию", "limit", options.Limit)
	}
	if options.Page <= 0 {
		options.Page = 1
		log.Debug("установлено значение page по умолчанию", "page", options.Page)
	}

	offset := (options.Page - 1) * options.Limit

	builder := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		OrderBy("date_time DESC").
		Limit(uint64(options.Limit)).
		Offset(uint64(offset))

	countBuilder := r.sb.Select("COUNT(*)").
		From("receptions")

	whereBuilder := squirrel.And{}

	if options.PVZID != uuid.Nil {
		whereBuilder = append(whereBuilder, squirrel.Eq{"pvz_id": options.PVZID})
		log.Debug("добавлен фильтр по ПВЗ", "pvz_id", options.PVZID)
	}

	if options.Status != "" {
		whereBuilder = append(whereBuilder, squirrel.Eq{"status": options.Status})
		log.Debug("добавлен фильтр по статусу", "status", options.Status)
	}

	if !options.FromDate.IsZero() {
		whereBuilder = append(whereBuilder, squirrel.GtOrEq{"date_time": options.FromDate})
		log.Debug("добавлен фильтр по начальной дате", "from_date", options.FromDate.Format(time.RFC3339))
	}

	if !options.ToDate.IsZero() {
		whereBuilder = append(whereBuilder, squirrel.LtOrEq{"date_time": options.ToDate})
		log.Debug("добавлен фильтр по конечной дате", "to_date", options.ToDate.Format(time.RFC3339))
	}

	if len(whereBuilder) > 0 {
		builder = builder.Where(whereBuilder)
		countBuilder = countBuilder.Where(whereBuilder)
	}

	sqlQuery, args, err := builder.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, 0, fmt.Errorf("error building SQL: %w", err)
	}

	if log.Enabled(ctx, logger.LevelDebug) {
		log.Debug("SQL запрос для списка приемок", "query", sqlQuery)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка выполнения запроса списка приемок", "error", err)
		return nil, 0, fmt.Errorf("error querying receptions: %w", err)
	}
	defer rows.Close()

	var receptions []*models.Reception
	for rows.Next() {
		var reception models.Reception
		if err := rows.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status); err != nil {
			log.Error("ошибка сканирования строки приемки", "error", err)
			return nil, 0, fmt.Errorf("error scanning reception row: %w", err)
		}
		receptions = append(receptions, &reception)
	}

	countSql, countArgs, err := countBuilder.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL для подсчета", "error", err)
		return nil, 0, fmt.Errorf("error building count SQL: %w", err)
	}

	var total int
	err = r.db.QueryRowContext(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		log.Error("ошибка подсчета общего количества приемок", "error", err)
		return nil, 0, fmt.Errorf("error counting total receptions: %w", err)
	}

	log.Info("список приемок успешно получен",
		"count", len(receptions),
		"total", total,
	)

	return receptions, total, nil
}

func (r *ReceptionRepository) GetReceptionWithProducts(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение приемки с товарами", "reception_id", id)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		log.Error("ошибка начала транзакции", "error", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		if err != nil {
			log.Debug("откат транзакции из-за ошибки")
			tx.Rollback()
		}
	}()

	receptionQuery := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{"id": id})

	receptionSql, receptionArgs, err := receptionQuery.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL для приемки", "error", err, "reception_id", id)
		return nil, fmt.Errorf("error building reception SQL: %w", err)
	}

	var reception models.Reception
	err = tx.QueryRowContext(ctx, receptionSql, receptionArgs...).Scan(
		&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("приемка не найдена", "reception_id", id)
			return nil, nil
		}
		log.Error("ошибка получения приемки", "error", err, "reception_id", id)
		return nil, fmt.Errorf("error getting reception by id: %w", err)
	}

	productsQuery := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"reception_id": id}).
		OrderBy("sequence_num")

	productsSql, productsArgs, err := productsQuery.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL для товаров", "error", err, "reception_id", id)
		return nil, fmt.Errorf("error building products SQL: %w", err)
	}

	rows, err := tx.QueryContext(ctx, productsSql, productsArgs...)
	if err != nil {
		log.Error("ошибка получения товаров для приемки", "error", err, "reception_id", id)
		return nil, fmt.Errorf("error querying products for reception: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum); err != nil {
			log.Error("ошибка сканирования строки товара", "error", err)
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, &product)
	}

	if err = tx.Commit(); err != nil {
		log.Error("ошибка фиксации транзакции", "error", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	reception.Products = products
	log.Info("приемка с товарами успешно получена",
		"reception_id", reception.ID,
		"pvz_id", reception.PVZID,
		"status", reception.Status,
		"products_count", len(products),
	)

	return &reception, nil
}
