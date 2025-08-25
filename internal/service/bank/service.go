package bank

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
	banks, err := s.repo.getBankList(ctx)
	if err != nil {
		logger.Error("failed to get bank list", err)
		return fmt.Errorf("service initialization failed: %w", err)
	}

	bankMap := make(map[int64]Bank, len(banks))
	for _, bank := range banks {
		bankMap[bank.ID] = bank
	}

	s.store.setBankList(banks)
	s.store.setBankMap(bankMap)

	return nil
}

func (s *Service) GetBank(ctx context.Context, id int64) (*Bank, error) {
	bank := s.store.getBank(id)
	if bank != nil {
		return bank, nil
	}

	logger.ErrorWithFields("failed to get bank from cache", nil, "bank_id", id)

	bank, err := s.repo.getBank(ctx, id)
	if err != nil {
		return nil, psql.MapPostgresError("failed to get bank", err)
	}

	return bank, nil
}

func (s *Service) BankList(ctx context.Context) ([]Bank, error) {
	banks := s.store.getBankList()
	if banks != nil {
		return banks, nil
	}

	logger.Error("failed to get bank list from cache", nil)

	banks, err := s.repo.getBankList(ctx)
	if err != nil {
		return nil, psql.MapPostgresError("failed to get banks", err)
	}

	return banks, nil
}
