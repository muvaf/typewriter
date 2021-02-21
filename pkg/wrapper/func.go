package wrapper

import (
	"fmt"
	"go/types"
	"strings"
)

func NewFunc() *Func {
	return &Func{}
}

type Func struct{}

func (f *Func) Wrap(a, b types.Type, content string) string {
	an := strings.Split(a.String(), "/")
	bn := strings.Split(b.String(), "/")
	return fmt.Sprintf("func lateInitialize(cr *%s, resp *%s) error {\n%s\nreturn nil\n}", an[len(an)-1], bn[len(bn)-1], content)
}