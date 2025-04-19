package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pvz-service/internal/domain/models"

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
	query := r.sb.Insert("pvz").
		Columns("city").
		Values(city).
		Suffix("RETURNING id, registration_date, city")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var pvz models.PVZ
	err = r.db.QueryRowContext(ctx, sql, args...).Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)

	if err != nil {
		return nil, fmt.Errorf("error creating PVZ: %w", err)
	}

	return &pvz, nil
}

func (r *PVZRepository) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	query := r.sb.Select("id", "registration_date", "city").
		From("pvz").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var pvz models.PVZ
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&pvz.ID, &pvz.RegistrationDate, &pvz.City,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting PVZ by id: %w", err)
	}

	return &pvz, nil
}

func (r *PVZRepository) ListPVZ(ctx context.Context, options models.PVZListOptions) ([]*models.PVZWithReceptionsResponse, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if options.Limit <= 0 {
		options.Limit = 10
	}
	if options.Page <= 0 {
		options.Page = 1
	}

	offset := (options.Page - 1) * options.Limit

	var pvzQuery squirrel.SelectBuilder
	var countQuery squirrel.SelectBuilder

	if !options.StartDate.IsZero() && !options.EndDate.IsZero() {
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
		pvzQuery = r.sb.Select("id", "registration_date", "city").
			From("pvz").
			OrderBy("id").
			Limit(uint64(options.Limit)).
			Offset(uint64(offset))

		countQuery = r.sb.Select("COUNT(*)").From("pvz")
	}

	pvzSql, pvzArgs, err := pvzQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building PVZ query: %w", err)
	}

	rows, err := tx.QueryContext(ctx, pvzSql, pvzArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying PVZ list: %w", err)
	}
	defer rows.Close()

	var pvzsWithReceptions []*models.PVZWithReceptionsResponse
	for rows.Next() {
		var pvz models.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			return nil, 0, fmt.Errorf("error scanning PVZ row: %w", err)
		}

		receptions, err := r.getReceptionsByPVZIDTx(ctx, tx, pvz.ID, options.StartDate, options.EndDate)
		if err != nil {
			return nil, 0, err
		}

		receptionWithProducts := make([]*models.ReceptionWithProducts, 0)
		for _, reception := range receptions {
			products, err := r.getProductsByReceptionIDTx(ctx, tx, reception.ID)
			if err != nil {
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
		return nil, 0, fmt.Errorf("error building count query: %w", err)
	}

	var total int
	err = tx.QueryRowContext(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting total PVZ: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return pvzsWithReceptions, total, nil
}

// Вспомогательный метод для получения приемок с помощью транзакции
func (r *PVZRepository) getReceptionsByPVZIDTx(ctx context.Context, tx *sql.Tx, pvzID uuid.UUID, startDate, endDate time.Time) ([]*models.Reception, error) {
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
		return nil, fmt.Errorf("error building receptions query: %w", err)
	}

	rows, err := tx.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting receptions for PVZ: %w", err)
	}
	defer rows.Close()

	var receptions []*models.Reception
	for rows.Next() {
		var reception models.Reception
		if err := rows.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status); err != nil {
			return nil, fmt.Errorf("error scanning reception row: %w", err)
		}
		receptions = append(receptions, &reception)
	}

	return receptions, nil
}

// Вспомогательный метод для получения товаров с помощью транзакции
func (r *PVZRepository) getProductsByReceptionIDTx(ctx context.Context, tx *sql.Tx, receptionID uuid.UUID) ([]*models.Product, error) {
	query := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("sequence_num")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building products query: %w", err)
	}

	rows, err := tx.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting products for reception: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID, &product.SequenceNum); err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}
