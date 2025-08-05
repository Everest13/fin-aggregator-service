package category

import (
	"context"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
	"github.com/Everest13/fin-aggregator-service/internal/utils/psql"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo  *repository
	store *Store
}

func NewService(dbPool *pgxpool.Pool) *Service {
	return &Service{
		repo:  newRepository(dbPool),
		store: NewStore(),
	}
}

func (s *Service) Initialize(ctx context.Context) error {
	keywords, err := s.repo.getCategoriesKeywords(ctx)
	if err != nil {
		logger.Error("failed to get category's keywords", err)
		return fmt.Errorf("service initialization failed: %w", err)
	}

	s.store.ReloadKeywordCategoryIDMap(keywords)

	categories, err := s.repo.categoryList(ctx)
	if err != nil {
		logger.Error("failed to get categories", err)
		return fmt.Errorf("service initialization failed: %w", err)
	}

	s.store.ReloadCategoriesMap(categories)

	return nil
}

func (s *Service) Store() *Store {
	return s.store
}

func (s *Service) CategoryList(ctx context.Context) ([]Category, error) {
	categories, err := s.repo.categoryList(ctx)
	if err != nil {
		logger.Error("failed to get categories", err)
		return nil, psql.MapPostgresError("failed to get categories", err)
	}

	return categories, nil
}

func (s *Service) GetCategoryByID(ctx context.Context, id int64) (*Category, error) {
	category := s.store.GetCategory(id)
	if category != nil {
		return category, nil
	}

	logger.ErrorWithFields("failed to get category from cache", nil, "category_id", id)

	category, err := s.repo.getCategoryByID(ctx, id)
	if err != nil {
		logger.ErrorWithFields("failed to get category", err, "category_id", id)
		return nil, psql.MapPostgresError("failed to get category", err)
	}

	return category, nil
}

func (s *Service) GetKeywordCategoryIDMap(ctx context.Context) (map[string]int64, error) {
	keywordCategoryMap := s.store.GetKeywordCategoryIDMap()
	if keywordCategoryMap != nil {
		return keywordCategoryMap, nil
	}

	logger.Error("failed to get keyword category's keywords from cache", nil)

	keywords, err := s.repo.getCategoriesKeywords(ctx)
	if err != nil {
		logger.Error("failed to get category's keywords", err)
		return nil, psql.MapPostgresError("failed to get category keywords", err)
	}

	for _, keyword := range keywords {
		keywordCategoryMap[keyword.Name] = keyword.CategoryID
	}

	return keywordCategoryMap, nil
}
