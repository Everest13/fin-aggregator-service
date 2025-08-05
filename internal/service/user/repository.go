package user

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

func (r *repository) getUserList(ctx context.Context) ([]User, error) {
	query, args, err := squirrel.
		Select(
			"u.id",
			"u.name",
			"ARRAY_REMOVE(ARRAY_AGG(ub.bank_id), NULL) AS banks",
		).
		From("users u").
		LeftJoin("user_bank ub ON u.id = ub.user_id").
		GroupBy("u.id, u.name").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var users []User
	err = pgxscan.Select(ctx, r.dbPool, &users, query, args...)
	if err != nil {
		return nil, err
	}

	return users, nil
}
