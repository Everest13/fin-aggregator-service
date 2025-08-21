package category

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

func (r *repository) categoryList(ctx context.Context) ([]Category, error) {
	query, args, err := squirrel.
		Select("*").
		From(categoryTable).
		OrderBy("name").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var categories []Category
	err = pgxscan.Select(ctx, r.dbPool, &categories, query, args...)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *repository) getCategoriesKeywords(ctx context.Context) ([]CategoryKeyword, error) {
	query, args, err := squirrel.
		Select("ck.id", "ck.category_id", "ck.name").
		From("category_keyword ck").
		Join("category c ON ck.category_id = c.id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var result []CategoryKeyword
	err = pgxscan.Select(ctx, r.dbPool, &result, query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *repository) getCategoryByID(ctx context.Context, categoryID int64) (*Category, error) {
	query, args, err := squirrel.
		Select("*").
		From(categoryTable).
		Where(squirrel.Eq{"id": categoryID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var category Category
	err = pgxscan.Get(ctx, r.dbPool, &category, query, args...)
	if err != nil {
		return nil, err
	}

	return &category, nil
}
