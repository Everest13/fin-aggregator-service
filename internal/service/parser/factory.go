package parser

import (
	"context"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"

	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
)

type parser interface {
	ParseRecords(ctx context.Context, records [][]string, targetFieldIds map[bank.TargetField][]int, bankID, userID int64) ([]*transaction.Transaction, map[int64][]error)
}

type parserFactory struct {
	parsers map[bank.BankName]parser
}

func newParserFactory(categoryService *category.Service) *parserFactory {
	createBase := func() *baseParser {
		bp := &baseParser{categoryService: categoryService}
		bp.initFieldFuncMap(bp)
		return bp
	}

	return &parserFactory{
		parsers: map[bank.BankName]parser{
			bank.AmexBankName:    newAmexParser(createBase()),
			bank.RevolutBankName: newRevolutParser(createBase()),
			bank.UnknownBankName: createBase(),
		},
	}
}

func (f *parserFactory) GetParser(bankName bank.BankName) parser {
	if p, exists := f.parsers[bankName]; exists {
		return p
	}

	logger.ErrorWithFields("parser not found", nil, "bank_name", bankName)
	return f.parsers[bank.UnknownBankName]
}
