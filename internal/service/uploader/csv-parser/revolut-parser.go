package csv_parser

import (
	"context"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"strconv"
)

type revolutParser struct {
	*BaseParser
}

func newRevolutParser(baseParser *BaseParser) *revolutParser {
	rP := &revolutParser{
		BaseParser: baseParser,
	}

	rP.initFieldFuncMap(rP)

	return rP
}

func (r *revolutParser) parseAmount(_ context.Context, tr *transaction.Transaction, data []string) error {
	if len(data) == 0 || data[0] == "" {
		return fmt.Errorf("empty amount data")
	}

	amountStr := data[0]
	switch amountStr[0] {
	case '-':
		tr.Type = transaction.OutcomeTransactionType
	default:
		tr.Type = transaction.IncomeTransactionType
	}

	if _, err := strconv.ParseFloat(amountStr, 64); err != nil {
		return fmt.Errorf("invalid amount format: %s", amountStr)
	}

	tr.Amount = amountStr
	return nil
}
