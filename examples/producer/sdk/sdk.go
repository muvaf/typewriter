package sdk

type SDKUser struct {
	Name       string
	Id         int
	Belongings []SDKBelonging
}

type SDKBelonging struct {
	Cars []string
}
