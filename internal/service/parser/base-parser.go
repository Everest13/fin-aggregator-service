package parser

import (
	"context"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"strconv"
	"strings"
	"time"
)

type baseParser struct {
	categoryService *category.Service
	fieldFuncMap    map[bank.TargetField]func(ctx context.Context, tr *transaction.Transaction, data []string) error
}

type fieldParser interface {
	parseDate(ctx context.Context, tr *transaction.Transaction, data []string) error
	parseAmount(ctx context.Context, tr *transaction.Transaction, data []string) error
	parseDescription(ctx context.Context, tr *transaction.Transaction, data []string) error
	parseCategory(ctx context.Context, tr *transaction.Transaction, data []string) error
	parseExternalID(ctx context.Context, tr *transaction.Transaction, data []string) error
}

func (p *baseParser) initFieldFuncMap(parser fieldParser) {
	p.fieldFuncMap = map[bank.TargetField]func(ctx context.Context, tr *transaction.Transaction, data []string) error{
		bank.DateTargetField:        parser.parseDate,
		bank.AmountTargetField:      parser.parseAmount,
		bank.DescriptionTargetField: parser.parseDescription,
		bank.CategoryTargetField:    parser.parseCategory,
		bank.ExternalIDTargetField:  parser.parseExternalID,
	}
}

func (p *baseParser) ParseRecords(ctx context.Context, records [][]string, targetFieldIds map[bank.TargetField][]int, bankID, userID int64) ([]*transaction.Transaction, map[int64][]error) {
	transactions := make([]*transaction.Transaction, 0, len(records)-1)

	processRecord := func(i int) (*transaction.Transaction, []error) {
		tr := &transaction.Transaction{
			BankID:     bankID,
			UserID:     userID,
			Type:       transaction.UnspecifiedTransactionType,
			CategoryID: category.UncategorizedID,
		}

		record := records[i]
		errs := []error{}
		for field, ids := range targetFieldIds {
			if len(ids) == 0 {
				continue
			}

			recordLen := len(records[i])
			if method, ok := p.fieldFuncMap[field]; ok {
				data := make([]string, 0, len(ids))
				for _, id := range ids {
					if id < recordLen {
						data = append(data, strings.TrimSpace(record[id]))
					} else {
						errs = append(errs, fmt.Errorf("missing field at index %d in row %d", id, i+1))
					}
				}

				err := method(ctx, tr, data)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}

		return tr, nil
	}

	recordErrs := map[int64][]error{}
	for i := 1; i < len(records); i++ {
		tr, errs := processRecord(i)
		if len(errs) != 0 {
			recordErrs[int64(i+1)] = errs
		}
		transactions = append(transactions, tr)
	}

	return transactions, recordErrs
}

var dateFormats = []string{
	"2006-01-02 15:04:05",
	"02/01/2006",
	"2006-01-02",
	"02.01.2006",
	"2006/01/02",
}

func (p *baseParser) parseDate(ctx context.Context, tr *transaction.Transaction, data []string) error {
	if len(data) == 0 || data[0] == "" {
		return fmt.Errorf("empty date data")
	}

	dateStr := data[0]
	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, dateStr); err == nil {
			tr.TransactionDate = t
			return nil
		}
	}

	return fmt.Errorf("unknown date format: %s", dateStr)
}

func (p *baseParser) parseCategory(ctx context.Context, tr *transaction.Transaction, data []string) error {
	keywordCategory, err := p.categoryService.GetKeywordCategoryIDMap(ctx)
	if err != nil {
		return err
	}

	combinedText := strings.ToLower(strings.Join(data, " "))
	for keyword, categoryID := range keywordCategory {
		if strings.Contains(combinedText, strings.ToLower(keyword)) {
			tr.CategoryID = categoryID
			return nil
		}
	}

	return nil
}

func (p *baseParser) parseAmount(_ context.Context, tr *transaction.Transaction, data []string) error {
	if len(data) == 0 || data[0] == "" {
		return fmt.Errorf("empty amount data")
	}

	amountStr := data[0]
	switch amountStr[0] {
	case '-':
		tr.Type = transaction.OutcomeTransactionType
	case '+':
		tr.Type = transaction.IncomeTransactionType
	default:
		tr.Type = transaction.UnspecifiedTransactionType
	}

	if _, err := strconv.ParseFloat(amountStr, 64); err != nil {
		return fmt.Errorf("invalid amount format: %s", amountStr)
	}

	tr.Amount = amountStr
	return nil
}

func (p *baseParser) parseDescription(ctx context.Context, tr *transaction.Transaction, data []string) error {
	s := data[0]
	for i := 1; i < len(data); i++ {
		s = s + ", " + data[i]
	}

	tr.Description = s

	return nil
}

func (p *baseParser) parseExternalID(ctx context.Context, tr *transaction.Transaction, data []string) error {
	tr.ExternalID = data[0]

	return nil
}
