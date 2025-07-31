package parser

import (
	"fmt"
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

func (a *amexParser) parseAmount(tr *transaction.Transaction, data []string) error {
	if len(data) == 0 || data[0] == "" {
		return fmt.Errorf("empty amount data")
	}

	tr.Type = transaction.UnspecifiedTransactionType

	amountStr := data[0]
	//todo check amount
	tr.Amount = amountStr

	return nil
}
