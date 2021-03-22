package app

// +typewriter:types:aggregated=github.com/muvaf/typewriter/examples/producer/sdk.SDKUser
type User struct {
	Name       string
	Id         int
	Belongings []Belonging
	UserGroup  string
}

type Belonging struct {
	Cars []string
}
