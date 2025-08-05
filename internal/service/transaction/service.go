package transaction

import (
	"context"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
	"github.com/Everest13/fin-aggregator-service/internal/utils/psql"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

type Service struct {
	repo            *repository
	categoryService *category.Service
}

func NewService(dbPool *pgxpool.Pool, categoryService *category.Service) *Service {
	repo := newRepository(dbPool)
	return &Service{
		repo:            repo,
		categoryService: categoryService,
	}
}

func (s *Service) Initialize(ctx context.Context) error {
	return s.PreCreatePartitions(ctx, 12)
}

// Creates parties for a year in advance
func (s *Service) PreCreatePartitions(ctx context.Context, months int) error {
	now := time.Now()

	for i := 0; i < months; i++ {
		monthDate := now.AddDate(0, i, 0)
		monthKey := monthDate.Format("2006_01")

		if err := s.repo.ensurePartition(ctx, monthKey); err != nil {
			logger.ErrorWithFields("failed to create partition", err, "month_key", monthKey)
			return fmt.Errorf("failed to create partition: %w", err)
		}
	}

	return nil
}

func (s *Service) GetSummaryTransactions(ctx context.Context, month int32, year int32, userID int64, bankID int64) (*TransactionSummary, error) {
	enrichedTrs, err := s.repo.enrichedTransactionList(ctx, month, year, userID, bankID)
	if err != nil {
		logger.ErrorWithFields("failed to get transactions", err, "user_id", userID, "bank_id", bankID, "month", month, "year", year)
		return nil, psql.MapPostgresError("failed to get transactions", err)
	}

	var totalIncome float64 = 0
	var totalOutcome float64 = 0
	for _, tr := range enrichedTrs {
		amount, err := strconv.ParseFloat(tr.Amount, 64)
		if err != nil {
			logger.ErrorWithFields("failed to parse transaction amount", err,
				"transaction_id", tr.ID,
				"external_id", tr.ExternalID,
				"amount", tr.Amount,
			)
			continue
		}

		switch tr.Type {
		case IncomeTransactionType:
			totalIncome += amount
		case OutcomeTransactionType:
			totalOutcome += amount
		}
	}

	return &TransactionSummary{
		Transactions: enrichedTrs,
		TotalCount:   len(enrichedTrs),
		TotalIncome:  strconv.FormatFloat(totalIncome, 'f', 2, 64),
		TotalOutcome: strconv.FormatFloat(totalOutcome, 'f', 2, 64),
	}, nil
}

func (s *Service) UpdateTransaction(ctx context.Context, data *TransactionUpdateData) (*EnrichedTransaction, error) {
	tr, err := s.repo.getEnrichedTransaction(ctx, data.ID)
	if err != nil {
		logger.ErrorWithFields("transaction not found", err, "transaction_id", data.ID)
		return nil, psql.MapPostgresError("transaction not found", err)
	}

	if data.Type != nil {
		tr.Type = *data.Type
	}

	if data.CategoryID != nil {
		ctgr, err := s.categoryService.GetCategoryByID(ctx, *data.CategoryID)
		if err != nil {
			return nil, err
		}
		tr.CategoryID = ctgr.ID
		tr.CategoryName = ctgr.Name
	}

	updatedTr, err := s.repo.updateTransaction(ctx, tr)
	if err != nil {
		logger.Error("failed to update transaction", err)
		return nil, psql.MapPostgresError("failed to update transaction", err)
	}

	newTr := &EnrichedTransaction{
		ID:              updatedTr.ID,
		ExternalID:      updatedTr.ExternalID,
		BankID:          updatedTr.BankID,
		UserID:          updatedTr.UserID,
		Amount:          updatedTr.Amount,
		CategoryID:      updatedTr.CategoryID,
		Description:     updatedTr.Description,
		Type:            updatedTr.Type,
		TransactionDate: updatedTr.TransactionDate,
		CreatedAt:       updatedTr.CreatedAt,
		BankName:        tr.BankName,
		CategoryName:    tr.CategoryName,
	}

	return newTr, nil
}

func (s *Service) SaveTransactions(ctx context.Context, transactions []*Transaction) error {
	if len(transactions) == 0 {
		return status.Errorf(codes.InvalidArgument, "no transactions to save")
	}

	err := s.repo.saveTransaction(ctx, transactions)
	if err != nil {
		return psql.MapPostgresError("failed to save transactions", err)
	}

	return err
}

func (s *Service) GetTransactionTypeList() []TransactionType {
	return []TransactionType{
		UnspecifiedTransactionType,
		IncomeTransactionType,
		OutcomeTransactionType,
	}
}
