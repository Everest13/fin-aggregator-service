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
		Select("*").
		From(userTable).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var users []User
	err = pgxscan.Select(ctx, r.dbPool, &users, query, args...)
	if err != nil {
		//todo
	}

	return users, nil
}
