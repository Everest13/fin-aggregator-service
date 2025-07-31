package transaction

import "time"

const (
	transactionTable = "transaction"
)

type TransactionType string

const (
	UnspecifiedTransactionType = "UNSPECIFIED"
	IncomeTransactionType      = "INCOME"
	OutcomeTransactionType     = "OUTCOME"
)

type Transaction struct {
	ID              int64
	ExternalID      string
	BankID          int64
	UserID          int64
	Amount          string
	CategoryID      int64
	Description     string
	Type            TransactionType
	TransactionDate time.Time
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

type EnrichedTransaction struct {
	ID              int64
	ExternalID      string
	BankID          int64
	UserID          int64
	Amount          string
	CategoryID      int64
	Description     string
	Type            TransactionType
	TransactionDate time.Time
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	BankName        string
	CategoryName    string
	UserName        string
}

type TransactionSummary struct {
	Transactions []EnrichedTransaction
	TotalCount   int
	TotalIncome  string
	TotalOutcome string
}

type TransactionUpdateData struct {
	ID         int64
	Type       *TransactionType
	CategoryID *int64
}
