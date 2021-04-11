package app

// +typewriter:types:aggregated=github.com/muvaf/typewriter/examples/producer/db.UserV1
// +typewriter:types:aggregated=github.com/muvaf/typewriter/examples/producer/db.UserV2
type UserAll struct {
	Name       string
	Surname    string
	Id         int
	Identifier int
	UserGroup  string
	Belongings []BelongingAll
}

type BelongingAll struct {
	Automobiles []string
	Cars        []string
}
