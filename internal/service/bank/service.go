package bank

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
	banks, err := s.repo.getBankList(ctx)
	if err != nil {
		return fmt.Errorf("service initialization failed: %w", err)
	}

	s.store.Reload(banks)

	return nil
}

func (s *Service) GetBank(ctx context.Context, id int64) (*Bank, error) {
	bank, err := s.repo.getBank(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get bank by ID %d: %v", id, err)
	}

	return bank, nil
}

func (s *Service) BankList(ctx context.Context) ([]Bank, error) {
	banks, err := s.repo.getBankList(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list banks: %v", err)
	}

	return banks, nil
}

func (s *Service) GetBankHeaders(ctx context.Context, bankID int64) ([]Header, error) {
	headers, err := s.repo.getBankHeaders(ctx, bankID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get headers for bank ID %d: %v", bankID, err)
	}

	return headers, nil
}
