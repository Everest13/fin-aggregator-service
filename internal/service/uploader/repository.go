package uploader

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"

	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	dbPool *pgxpool.Pool
}

func newRepository(dbPool *pgxpool.Pool) *repository {
	return &repository{
		dbPool: dbPool,
	}
}

func (r *repository) getBankHeaderMappings(ctx context.Context) ([]HeaderMapping, error) {
	query, args, err := squirrel.
		Select(
			"bh.bank_id",
			"bh.name",
			"bh.required",
			"array_agg(bhm.transaction_field) AS tr_fields",
		).
		From("bank_header bh").
		Join("bank_header_mapping bhm ON bh.id = bhm.header_id").
		GroupBy("bh.id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var headerMapping []HeaderMapping
	if err = pgxscan.Select(ctx, r.dbPool, &headerMapping, query, args...); err != nil {
		return nil, err
	}

	return headerMapping, nil
}

func (r *repository) getBankHeaderMappingsByBank(ctx context.Context, bankID int64) ([]HeaderMapping, error) {
	query, args, err := squirrel.
		Select(
			"bh.bank_id",
			"bh.name",
			"bh.required",
			"array_agg(bhm.transaction_field) AS tr_fields",
		).
		From("bank_header bh").
		Join("bank_header_mapping bhm ON bh.id = bhm.header_id").
		Where(squirrel.Eq{"bh.bank_id": bankID}).
		GroupBy("bh.id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var headerMapping []HeaderMapping
	if err = pgxscan.Select(ctx, r.dbPool, &headerMapping, query, args...); err != nil {
		return nil, err
	}

	return headerMapping, nil
}
