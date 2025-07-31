package user

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	repo *repository
}

func NewService(dbPool *pgxpool.Pool) *Service {
	return &Service{
		repo: newRepository(dbPool),
	}
}

func (s *Service) UserList(ctx context.Context) ([]User, error) {
	banks, err := s.repo.getUserList(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user list: %v", err)
	}

	return banks, nil
}
