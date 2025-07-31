package transaction

import (
	"context"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
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

// todo Создаём партиции на год вперёд
func (s *Service) Initialize(ctx context.Context) error {
	return s.PreCreatePartitions(ctx, 12)
}

func (s *Service) PreCreatePartitions(ctx context.Context, months int) error {
	now := time.Now()

	for i := 0; i < months; i++ {
		monthDate := now.AddDate(0, i, 0)
		monthKey := monthDate.Format("2006_01")

		if err := s.repo.ensurePartition(ctx, monthKey); err != nil {
			//todo log + err
			log.Printf("Failed to create partition for %s: %v", monthKey, err)
		}
	}

	return nil
}

func (s *Service) GetTransactions(ctx context.Context, month int32, year int32, userID int64, bankID int64) (*TransactionSummary, error) {
	enrichedTrs, err := s.repo.getTransactions(ctx, month, year, userID, bankID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	var totalIncome float64 = 0
	var totalOutcome float64 = 0
	for _, tr := range enrichedTrs {
		amount, err := strconv.ParseFloat(tr.Amount, 64)
		if err != nil {
			//todo
			// игнорируем или логируем ошибку парсинга суммы
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

// todo переделать на работу с обычной транзакции которую можно без join получить
func (s *Service) UpdateTransaction(ctx context.Context, data *TransactionUpdateData) (*EnrichedTransaction, error) {
	enrichedTr, err := s.repo.getEnrichedTransaction(ctx, data.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "transaction with ID %s not found: %v", data.ID, err)
	}

	if data.Type != nil {
		enrichedTr.Type = *data.Type
	}

	if data.CategoryID != nil {
		category, err := s.categoryService.GetCategoryByID(ctx, *data.CategoryID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "category with ID %s not found: %v", *data.CategoryID, err)
		}
		enrichedTr.CategoryID = category.ID
		enrichedTr.CategoryName = category.Name
	}

	updatedTr, err := s.repo.updateTransaction(ctx, enrichedTr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update transaction: %v", err)
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
		BankName:        enrichedTr.BankName,
		CategoryName:    enrichedTr.CategoryName,
	}

	return newTr, nil
}

func (s *Service) SaveTransactions(ctx context.Context, transactions []*Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return s.repo.saveTransaction(ctx, transactions)
}
