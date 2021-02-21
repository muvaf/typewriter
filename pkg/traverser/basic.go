package typewriter

import (
	"fmt"
	"go/types"
	"strings"
)

const (
	errFmtNotSameKind = "not same basic kind: %s and %s"
	errFmtUnknownKind = "unknown basic kind: %s"
)

func WithAssignmentTmpl(kind types.BasicKind, tmpl string) BasicOption {
	return func(b *Basic) {
		b.Templates[kind] = tmpl
	}
}

func WithDefaultAssignmentTmpl(tmpl string) BasicOption {
	return func(b *Basic) {
		for i := 1; i < 26; i++ {
			b.Templates[types.BasicKind(i)] = tmpl
		}
	}
}

type BasicOption func(*Basic)

const (
	DefaultAssignmentTmpl = "$a = $b"
)

func NewBasic(opts ...BasicOption) *Basic {
	b := &Basic{
		Templates: map[types.BasicKind]string{},
	}
	for i := 1; i < 26; i++ {
		b.Templates[types.BasicKind(i)] = DefaultAssignmentTmpl
	}
	for _, f := range opts {
		f(b)
	}
	return b
}

type Basic struct {
	Templates map[types.BasicKind]string
}

func (s *Basic) Print(a, b *types.Basic, aFieldPath, bFieldPath string) (string, error) {
	if a.Kind() != b.Kind() {
		return "", fmt.Errorf(errFmtNotSameKind, a.String(), b.String())
	}
	result, ok := s.Templates[a.Kind()]
	if !ok {
		return "", fmt.Errorf(errFmtUnknownKind, a.String())
	}
	result = strings.ReplaceAll(result, "$a", aFieldPath)
	result = strings.ReplaceAll(result, "$b", bFieldPath)
	return result + "\n", nil
}
