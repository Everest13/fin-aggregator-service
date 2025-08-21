package bank

import (
	"context"
	"github.com/Everest13/fin-aggregator-service/internal/utils/psql"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *repository
}

func NewService(dbPool *pgxpool.Pool) *Service {
	return &Service{
		repo: newRepository(dbPool),
	}
}

func (s *Service) GetBank(ctx context.Context, id int64) (*Bank, error) {
	bank, err := s.repo.getBank(ctx, id)
	if err != nil {
		return nil, psql.MapPostgresError("failed to get bank", err)
	}

	return bank, nil
}

func (s *Service) BankList(ctx context.Context) ([]Bank, error) {
	banks, err := s.repo.getBankList(ctx)
	if err != nil {
		return nil, psql.MapPostgresError("failed to get banks", err)
	}

	return banks, nil
}
