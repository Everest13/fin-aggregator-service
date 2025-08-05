package parser

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
)

type amexParser struct {
	*baseParser
}

func newAmexParser(baseParser *baseParser) *amexParser {
	aP := &amexParser{
		baseParser: baseParser,
	}

	aP.initFieldFuncMap(aP)

	return aP
}

func (a *amexParser) parseAmount(_ context.Context, tr *transaction.Transaction, data []string) error {
	if len(data) == 0 || data[0] == "" {
		return fmt.Errorf("empty amount data")
	}

	tr.Type = transaction.UnspecifiedTransactionType

	amountStr := data[0]

	if _, err := strconv.ParseFloat(amountStr, 64); err != nil {
		return fmt.Errorf("invalid amount format: %s", amountStr)
	}

	tr.Amount = amountStr
	return nil
}
