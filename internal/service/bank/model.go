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

type ImportMethod string

const (
	UndefinedImportMethod ImportMethod = "UNDEFINED"
	CSVImportMethod       ImportMethod = "CSV"
	APIImportMethod       ImportMethod = "API"
)

type Bank struct {
	ID           int64
	Name         string
	ImportMethod ImportMethod
}

type BankHeader struct {
	ID       int64
	BankID   int64
	Name     string
	Required bool
}
