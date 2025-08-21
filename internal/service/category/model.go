package category

const categoryTable = "category"

const UncategorizedID = int64(1)

const TransferCategoryName = "Transfer"

type Category struct {
	ID          int64
	Name        string
	Description *string
}

type CategoryKeyword struct {
	ID         int64
	CategoryID int64
	Name       string
}
