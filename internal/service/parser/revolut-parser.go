package parser

import (
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
)

type revolutParser struct {
	*baseParser
}

func newRevolutParser(baseParser *baseParser) *revolutParser {
	rP := &revolutParser{
		baseParser: baseParser,
	}

	rP.baseParser.fieldFuncMap[bank.AmountTargetField] = rP.parseAmount

	return rP
}

func (r *revolutParser) parseAmount(tr *transaction.Transaction, data []string) error {
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

	//todo check amount
	tr.Amount = amountStr

	return nil
}
