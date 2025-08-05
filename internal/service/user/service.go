package user

import (
	"context"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
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
	users, err := s.repo.getUserList(ctx)
	if err != nil {
		logger.Error("failed to get user list", err)
		return nil, status.Errorf(codes.Internal, "failed to get user list")
	}

	return users, nil
}
