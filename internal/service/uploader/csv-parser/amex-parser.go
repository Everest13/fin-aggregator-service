package csv_parser

import (
	"context"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"strconv"
	"strings"

	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
)

type amexParser struct {
	*BaseParser
}

func newAmexParser(baseParser *BaseParser) *amexParser {
	aP := &amexParser{
		BaseParser: baseParser,
	}

	aP.initFieldFuncMap(aP)

	return aP
}

func (a *amexParser) parseAmount(_ context.Context, tr *transaction.Transaction, data []string) error {
	if len(data) == 0 || data[0] == "" {
		return fmt.Errorf("empty amount data")
	}

	amountStr := data[0]

	if _, err := strconv.ParseFloat(amountStr, 64); err != nil {
		return fmt.Errorf("invalid amount format: %s", amountStr)
	}

	tr.Amount = amountStr
	return nil
}

func (a *amexParser) parseCategory(ctx context.Context, tr *transaction.Transaction, data []string) error {
	keywordCategory, err := a.categoryService.GetKeywordCategoryIDMap(ctx)
	if err != nil {
		return err
	}

	combinedText := strings.ToLower(strings.Join(data, " "))
	for keyword, categoryID := range keywordCategory {
		if strings.Contains(combinedText, strings.ToLower(keyword)) {
			tr.CategoryID = categoryID
			break
		}
	}

	if tr.CategoryID == category.UncategorizedID {
		return nil
	}

	ctry, err := a.categoryService.GetCategoryByID(ctx, tr.CategoryID)
	if err != nil {
		return err
	}

	if ctry.Name != category.TransferCategoryName {
		tr.Type = transaction.OutcomeTransactionType
	}

	return nil
}
