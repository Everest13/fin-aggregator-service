package uploader

import (
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
)

type HeaderMapping struct {
	BankID   int64
	Name     string
	Required bool
	TrFields []transaction.TransactionField
}
