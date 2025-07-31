package category

const categoryTable = "category"

const UncategorizedID = int64(1)

type Category struct {
	ID          int64
	Name        string
	Description *string
}

type Keyword struct {
	ID         int64
	CategoryID int64
	Name       string
}
