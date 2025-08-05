package user

const userTable = "users"

type User struct {
	ID    int64
	Name  string
	Banks []int64
}
