package monzo

import (
	"context"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"strings"
	"time"
)

var dateFormats = []string{
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05.00Z",
	"2006-01-02T15:04:05.0Z",
	"2006-01-02T15:04:05Z",
}

func (s *Service) parseMonzoTransactions(
	ctx context.Context,
	monzoTransactions []MonzoTransaction,
	since time.Time,
	userID, bankID int64,
) ([]*transaction.Transaction, map[string][]MonzoTransaction, error) {
	trs := make([]*transaction.Transaction, 0, len(monzoTransactions))
	failedTrErr := map[string][]MonzoTransaction{}

	for _, mTr := range monzoTransactions {
		date, err := parseDate(mTr.Created, since)
		if err != nil {
			errStr := err.Error()
			failedTrErr[errStr] = append(failedTrErr[errStr], mTr)
			continue
		}

		ctgID, err := s.parseCategory(ctx, mTr.Category, mTr.Description)
		if err != nil {
			errStr := err.Error()
			failedTrErr[errStr] = append(failedTrErr[errStr], mTr)
		}

		tr := &transaction.Transaction{
			UserID:          userID,
			BankID:          bankID,
			ExternalID:      mTr.ID,
			Amount:          parseAmount(mTr.Amount),
			Description:     parseDescription(mTr.Description, mTr.Category, mTr.Notes, mTr.Scheme),
			TransactionDate: date,
			CategoryID:      ctgID,
			Type:            parseType(mTr.Amount),
		}

		trs = append(trs, tr)
	}

	return trs, failedTrErr, nil
}

func parseAmount(amount int64) string {
	return fmt.Sprintf("%.2f", float64(amount)/100)
}

func parseType(amount int64) transaction.TransactionType {
	if amount < 0 {
		return transaction.OutcomeTransactionType
	}
	return transaction.IncomeTransactionType
}

func parseDescription(desc, category, notes, scheme string) string {
	parts := []string{}

	if desc = strings.TrimSpace(desc); desc != "" {
		parts = append(parts, desc)
	}
	if category = strings.TrimSpace(category); category != "" {
		parts = append(parts, category)
	}
	if notes = strings.TrimSpace(notes); notes != "" {
		parts = append(parts, notes)
	}
	if scheme = strings.TrimSpace(scheme); scheme != "" {
		parts = append(parts, scheme)
	}

	return strings.Join(parts, ", ")
}

func parseDate(createdAt string, from time.Time) (time.Time, error) {
	if createdAt == "" {
		return from, fmt.Errorf("empty date data")
	}

	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, createdAt); err == nil {
			return t, nil
		}
	}

	return from, fmt.Errorf("unknown date format: %s", createdAt)
}

func (s *Service) parseCategory(ctx context.Context, mCategory, desc string) (int64, error) {
	keywordCategory, err := s.categoryService.GetKeywordCategoryIDMap(ctx)
	if err != nil {
		return category.UncategorizedID, err
	}

	combinedText := strings.ToLower(strings.Join([]string{mCategory, desc}, " "))
	for keyword, categoryID := range keywordCategory {
		if strings.Contains(combinedText, strings.ToLower(keyword)) {
			return categoryID, nil
		}
	}

	return category.UncategorizedID, nil
}
