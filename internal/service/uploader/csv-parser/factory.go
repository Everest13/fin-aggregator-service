package csv_parser

import (
	"context"

	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
)

type Parser interface {
	ParseRecords(ctx context.Context, records [][]string, targetFieldIds map[transaction.TransactionField][]int, bankID, userID int64) ([]*transaction.Transaction, map[int64][]error)
}

type Factory struct {
	parsers map[bank.BankName]Parser
}

func NewFactory(categoryService *category.Service) *Factory {
	createBase := func() *BaseParser {
		bp := &BaseParser{categoryService: categoryService}
		bp.initFieldFuncMap(bp)
		return bp
	}

	return &Factory{
		parsers: map[bank.BankName]Parser{
			bank.AmexBankName:    newAmexParser(createBase()),
			bank.RevolutBankName: newRevolutParser(createBase()),
			bank.UnknownBankName: createBase(),
		},
	}
}

func (f *Factory) GetParser(bankName bank.BankName) Parser {
	if p, exists := f.parsers[bankName]; exists {
		return p
	}

	logger.ErrorWithFields("uploader not found", nil, "bank_name", bankName)
	return f.parsers[bank.UnknownBankName]
}
