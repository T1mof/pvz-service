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
	query := r.sb.Insert("receptions").
		Columns("pvz_id", "status").
		Values(pvzID, models.StatusInProgress).
		Suffix("RETURNING id, date_time, pvz_id, status")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var reception models.Reception
	err = r.db.QueryRowContext(ctx, sql, args...).Scan(
		&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating reception: %w", err)
	}

	return &reception, nil
}

func (r *ReceptionRepository) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	query := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var reception models.Reception
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting reception by id: %w", err)
	}

	return &reception, nil
}

func (r *ReceptionRepository) GetLastOpenReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
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
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var reception models.Reception
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting last open reception: %w", err)
	}

	return &reception, nil
}

func (r *ReceptionRepository) CloseReception(ctx context.Context, id uuid.UUID) error {
	query := r.sb.Update("receptions").
		Set("status", models.StatusClosed).
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error closing reception: %w", err)
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
	if options.Limit <= 0 {
		options.Limit = 10
	}
	if options.Page <= 0 {
		options.Page = 1
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
	}

	if options.Status != "" {
		whereBuilder = append(whereBuilder, squirrel.Eq{"status": options.Status})
	}

	if !options.FromDate.IsZero() {
		whereBuilder = append(whereBuilder, squirrel.GtOrEq{"date_time": options.FromDate})
	}

	if !options.ToDate.IsZero() {
		whereBuilder = append(whereBuilder, squirrel.LtOrEq{"date_time": options.ToDate})
	}

	if len(whereBuilder) > 0 {
		builder = builder.Where(whereBuilder)
		countBuilder = countBuilder.Where(whereBuilder)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying receptions: %w", err)
	}
	defer rows.Close()

	var receptions []*models.Reception
	for rows.Next() {
		var reception models.Reception
		if err := rows.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status); err != nil {
			return nil, 0, fmt.Errorf("error scanning reception row: %w", err)
		}
		receptions = append(receptions, &reception)
	}

	countSql, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building count SQL: %w", err)
	}

	var total int
	err = r.db.QueryRowContext(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting total receptions: %w", err)
	}

	return receptions, total, nil
}

func (r *ReceptionRepository) GetReceptionWithProducts(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	receptionQuery := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{"id": id})

	receptionSql, receptionArgs, err := receptionQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building reception SQL: %w", err)
	}

	var reception models.Reception
	err = tx.QueryRowContext(ctx, receptionSql, receptionArgs...).Scan(
		&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting reception by id: %w", err)
	}

	productsQuery := r.sb.Select("id", "date_time", "type", "reception_id", "sequence_num").
		From("products").
		Where(squirrel.Eq{"reception_id": id}).
		OrderBy("sequence_num")

	productsSql, productsArgs, err := productsQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building products SQL: %w", err)
	}

	rows, err := tx.QueryContext(ctx, productsSql, productsArgs...)
	if err != nil {
		return nil, fmt.Errorf("error querying products for reception: %w", err)
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

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	reception.Products = products
	return &reception, nil
}
