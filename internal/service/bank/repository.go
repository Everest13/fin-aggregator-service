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
		Select(
			"b.id",
			"b.name",
			"COALESCE(array_agg(bim.import_method) FILTER (WHERE bim.import_method IS NOT NULL), '{}') AS import_methods",
		).
		From("bank b").
		LeftJoin("bank_import_method bim ON b.id = bim.bank_id").
		Where(squirrel.Eq{"b.id": id}).
		GroupBy("b.id", "b.name").
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
		Select(
			"b.id",
			"b.name",
			"COALESCE(array_agg(bim.import_method) FILTER (WHERE bim.import_method IS NOT NULL), '{}') AS import_methods",
		).
		From("bank b").
		LeftJoin("bank_import_method bim ON b.id = bim.bank_id").
		GroupBy("b.id").
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
