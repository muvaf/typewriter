package db

type UserV2 struct {
	Name       string
	Id         int
	Belongings []BelongingV2
	UserGroup  string
}

type BelongingV2 struct {
	Cars []string
}
