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

type PVZRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

func NewPVZRepository(db *sql.DB) *PVZRepository {
	return &PVZRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *PVZRepository) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	log := logger.FromContext(ctx)
	log.Debug("создание ПВЗ", "city", city)

	query := r.sb.Insert("pvz").
		Columns("city").
		Values(city).
		Suffix("RETURNING id, registration_date, city")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	if log.Enabled(ctx, logger.LevelDebug) {
		log.Debug("SQL запрос", "query", sqlQuery, "args", args)
	}

	var pvz models.PVZ
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)

	if err != nil {
		log.Error("ошибка создания ПВЗ в БД", "error", err, "city", city)
		return nil, fmt.Errorf("error creating PVZ: %w", err)
	}

	log.Info("ПВЗ успешно создан", "pvz_id", pvz.ID, "city", pvz.City)
	return &pvz, nil
}

func (r *PVZRepository) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение ПВЗ по ID", "pvz_id", id)

	query := r.sb.Select("id", "registration_date", "city").
		From("pvz").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err, "pvz_id", id)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var pvz models.PVZ
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&pvz.ID, &pvz.RegistrationDate, &pvz.City,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("ПВЗ не найден", "pvz_id", id)
			return nil, nil
		}
		log.Error("ошибка получения ПВЗ", "error", err, "pvz_id", id)
		return nil, fmt.Errorf("error getting PVZ by id: %w", err)
	}

	log.Debug("ПВЗ успешно получен", "pvz_id", pvz.ID, "city", pvz.City)
	return &pvz, nil
}

func (r *PVZRepository) ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение списка ПВЗ",
		"page", options.Page,
		"limit", options.Limit,
		"has_start_date", !options.StartDate.IsZero(),
		"has_end_date", !options.EndDate.IsZero(),
	)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error("ошибка начала транзакции", "error", err)
		return nil, 0, fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		if err != nil {
			log.Debug("откат транзакции из-за ошибки")
			tx.Rollback()
		}
	}()

	if options.Limit <= 0 {
		options.Limit = 10
		log.Debug("установлено значение limit по умолчанию", "limit", options.Limit)
	}
	if options.Page <= 0 {
		options.Page = 1
		log.Debug("установлено значение page по умолчанию", "page", options.Page)
	}

	offset := (options.Page - 1) * options.Limit

	var pvzQuery squirrel.SelectBuilder
	var countQuery squirrel.SelectBuilder

	if !options.StartDate.IsZero() && !options.EndDate.IsZero() {
		log.Debug("применение фильтра по датам",
			"start_date", options.StartDate.Format(time.RFC3339),
			"end_date", options.EndDate.Format(time.RFC3339),
		)

		pvzQuery = r.sb.Select("DISTINCT p.id", "p.registration_date", "p.city").
			From("pvz p").
			Join("receptions r ON p.id = r.pvz_id").
			Where(squirrel.And{
				squirrel.GtOrEq{"r.date_time": options.StartDate},
				squirrel.LtOrEq{"r.date_time": options.EndDate},
			}).
			OrderBy("p.id").
			Limit(uint64(options.Limit)).
			Offset(uint64(offset))

		countQuery = r.sb.Select("COUNT(DISTINCT p.id)").
			From("pvz p").
			Join("receptions r ON p.id = r.pvz_id").
			Where(squirrel.And{
				squirrel.GtOrEq{"r.date_time": options.StartDate},
				squirrel.LtOrEq{"r.date_time": options.EndDate},
			})
	} else {
		log.Debug("получение всех ПВЗ без фильтра по датам")

		pvzQuery = r.sb.Select("id", "registration_date", "city").
			From("pvz").
			OrderBy("id").
			Limit(uint64(options.Limit)).
			Offset(uint64(offset))

		countQuery = r.sb.Select("COUNT(*)").From("pvz")
	}

	pvzSql, pvzArgs, err := pvzQuery.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL для списка ПВЗ", "error", err)
		return nil, 0, fmt.Errorf("error building PVZ query: %w", err)
	}

	if log.Enabled(ctx, logger.LevelDebug) {
		log.Debug("SQL запрос для списка ПВЗ", "query", pvzSql)
	}

	rows, err := tx.QueryContext(ctx, pvzSql, pvzArgs...)
	if err != nil {
		log.Error("ошибка выполнения запроса списка ПВЗ", "error", err)
		return nil, 0, fmt.Errorf("error querying PVZ list: %w", err)
	}
	defer rows.Close()

	var pvzsWithReceptions []*models.PVZWithReceptionsResponse
	for rows.Next() {
		var pvz models.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			log.Error("ошибка сканирования строки ПВЗ", "error", err)
			return nil, 0, fmt.Errorf("error scanning PVZ row: %w", err)
		}

		log.Debug("получение приемок для ПВЗ", "pvz_id", pvz.ID)
		receptions, err := r.getReceptionsByPVZIDTx(ctx, tx, pvz.ID, options.StartDate, options.EndDate)
		if err != nil {
			log.Error("ошибка получения приемок для ПВЗ", "error", err, "pvz_id", pvz.ID)
			return nil, 0, err
		}

		receptionWithProducts := make([]*models.ReceptionWithProducts, 0)
		for _, reception := range receptions {
			log.Debug("получение товаров для приемки", "reception_id", reception.ID)
			products, err := r.getProductsByReceptionIDTx(ctx, tx, reception.ID)
			if err != nil {
				log.Error("ошибка получения товаров для приемки",
					"error", err,
					"reception_id", reception.ID,
				)
				return nil, 0, err
			}

			receptionWithProducts = append(receptionWithProducts, &models.ReceptionWithProducts{
				Reception: reception,
				Products:  products,
			})
		}

		pvzsWithReceptions = append(pvzsWithReceptions, &models.PVZWithReceptionsResponse{
			PVZ:        &pvz,
			Receptions: receptionWithProducts,
		})
	}

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL для подсчета ПВЗ", "error", err)
		return nil, 0, fmt.Errorf("error building count query: %w", err)
	}

	var total int
	err = tx.QueryRowContext(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		log.Error("ошибка подсчета общего количества ПВЗ", "error", err)
		return nil, 0, fmt.Errorf("error counting total PVZ: %w", err)
	}

	if err = tx.Commit(); err != nil {
		log.Error("ошибка фиксации транзакции", "error", err)
		return nil, 0, fmt.Errorf("error committing transaction: %w", err)
	}

	log.Info("список ПВЗ успешно получен",
		"count", len(pvzsWithReceptions),
		"total", total,
	)

	return pvzsWithReceptions, total, nil
}

func (r *PVZRepository) getReceptionsByPVZIDTx(ctx context.Context, tx *sql.Tx, pvzID uuid.UUID, startDate, endDate time.Time) ([]*models.Reception, error) {
	log := logger.FromContext(ctx)

	var query squirrel.SelectBuilder

	if !startDate.IsZero() && !endDate.IsZero() {
		query = r.sb.Select("id", "date_time", "pvz_id", "status").
			From("receptions").
			Where(squirrel.And{
				squirrel.Eq{"pvz_id": pvzID},
				squirrel.GtOrEq{"date_time": startDate},
				squirrel.LtOrEq{"date_time": endDate},
			}).
			OrderBy("date_time")
	} else {
		query = r.sb.Select("id", "date_time", "pvz_id", "status").
			From("receptions").
			Where(squirrel.Eq{"pvz_id": pvzID}).
			OrderBy("date_time")
	}

	sql, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL для приемок", "error", err, "pvz_id", pvzID)
		return nil, fmt.Errorf("error building receptions query: %w", err)
	}

	rows, err := tx.QueryContext(ctx, sql, args...)
	if err != nil {
		log.Error("ошибка получения приемок для ПВЗ", "error", err, "pvz_id", pvzID)
		return nil, fmt.Errorf("error getting receptions for PVZ: %w", err)
	}
	defer rows.Close()

	var receptions []*models.Reception
	for rows.Next() {
		var reception models.Reception
		if err := rows.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status); err != nil {
			log.Error("ошибка сканирования строки приемки", "error", err)
			return nil, fmt.Errorf("error scanning reception row: %w", err)
		}
		receptions = append(receptions, &reception)
	}

	log.Debug("получены приемки для ПВЗ", "pvz_id", pvzID, "count", len(receptions))
	return receptions, nil
}

func (r *PVZRepository) getProductsByReceptionIDTx(ctx context.Context, tx *sql.Tx, receptionID uuid.UUID) ([]*models.Product, error) {
	log := logger.FromContext(ctx)

	query := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("sequence_num")

	sql, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL для товаров", "error", err, "reception_id", receptionID)
		return nil, fmt.Errorf("error building products query: %w", err)
	}

	rows, err := tx.QueryContext(ctx, sql, args...)
	if err != nil {
		log.Error("ошибка получения товаров для приемки", "error", err, "reception_id", receptionID)
		return nil, fmt.Errorf("error getting products for reception: %w", err)
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

	log.Debug("получены товары для приемки", "reception_id", receptionID, "count", len(products))
	return products, nil
}
