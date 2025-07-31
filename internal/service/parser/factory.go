package parser

import (
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
)

type parser interface {
	ParseRecords(records [][]string, targetFieldIds map[bank.TargetField][]int, bankID, userID int64) ([]*transaction.Transaction, map[int64][]error)
}

type parserFactory struct {
	parsers map[bank.BankName]parser
}

func newParserFactory(categoryStore *category.Store) *parserFactory {
	createBase := func() *baseParser {
		bp := &baseParser{categoryStore: categoryStore}
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

	return f.parsers[bank.UnknownBankName]
}
