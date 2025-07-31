package category

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	categories, err := s.repo.getCategoriesKeywords(ctx)
	if err != nil {
		return fmt.Errorf("service initialization failed: %w", err)
	}

	s.store.Reload(categories)

	return nil
}

func (s *Service) Store() *Store {
	return s.store
}

func (s *Service) CategoryList(ctx context.Context) ([]Category, error) {
	categories, err := s.repo.categoryList(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get category list: %v", err)
	}

	return categories, nil
}

func (s *Service) GetCategoryByID(ctx context.Context, id int64) (*Category, error) {
	category, err := s.repo.getCategoryByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get category list: %v", err)
	}

	return category, nil
}
