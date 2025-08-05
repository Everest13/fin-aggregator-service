package bank

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

var queryBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type repository struct {
	dbPool *pgxpool.Pool
}

func newRepository(dbPool *pgxpool.Pool) *repository {
	return &repository{
		dbPool: dbPool,
	}
}

func (r *repository) getBank(ctx context.Context, id int64) (*Bank, error) {
	query, args, err := squirrel.
		Select("*").
		From(bankTable).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var bank Bank
	err = pgxscan.Get(ctx, r.dbPool, &bank, query, args...)
	if err != nil {
		return nil, err
	}

	return &bank, nil
}

func (r *repository) getBankList(ctx context.Context) ([]Bank, error) {
	query, args, err := squirrel.
		Select("*").
		From(bankTable).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var banks []Bank
	err = pgxscan.Select(ctx, r.dbPool, &banks, query, args...)
	if err != nil {
		return nil, err
	}

	return banks, nil
}

func (r *repository) getBankHeaders(ctx context.Context, bankID int64) ([]Header, error) {
	query, args, err := squirrel.
		Select(
			"id",
			"bank_id",
			"name",
			"ARRAY_AGG(target_field) AS target_field",
		).
		From(bankHeaderTable).
		Where(squirrel.Eq{"bank_id": bankID}).
		GroupBy("id", "bank_id", "name").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var headers []Header
	err = pgxscan.Select(ctx, r.dbPool, &headers, query, args...)
	if err != nil {
		return nil, err
	}

	return headers, nil
}
