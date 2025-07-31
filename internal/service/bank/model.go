package bank

const (
	bankTable       = "bank"
	bankHeaderTable = "bank_header"
)

type BankName string

const (
	AmexBankName    BankName = "American express"
	RevolutBankName BankName = "Revolut"
	UnknownBankName BankName = "Unknown"
)

type TargetField string

const (
	DateTargetField        TargetField = "DATE"
	AmountTargetField      TargetField = "AMOUNT"
	CategoryTargetField    TargetField = "CATEGORY"
	ExternalIDTargetField  TargetField = "EXTERNALID"
	DescriptionTargetField TargetField = "DESCRIPTION"
)

type Bank struct {
	ID   int64
	Name string
}

// todo отказаться от []TargetField и работать как с категориями keywords
type Header struct {
	ID          int64
	BankID      int64
	Name        string
	TargetField []TargetField
}
