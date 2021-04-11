package db

type UserV1 struct {
	Name       string
	Surname    string
	Identifier int
	Belongings []BelongingV1
}

type BelongingV1 struct {
	Automobiles []string
}
