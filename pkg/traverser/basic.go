package typewriter

import (
	"fmt"
	"go/types"
	"strings"
)

type BasicOption func(*Basic)

const (
	DefaultFmtStringAssignment = "$a = xpv1.LateInitializeString($a, $b)"
	DefaultFmtBoolAssignment   = "$a = xpv1.LateInitializeBool($a, $b)"
	DefaultFmtInt64Assignment  = "$a = xpv1.LateInitializeInt64($a, $b)"
)

func NewBasic(opts ...BasicOption) *Basic {
	b := &Basic{
		Templates: map[types.BasicKind]string{
			types.String: DefaultFmtStringAssignment,
			types.Bool:   DefaultFmtBoolAssignment,
			types.Int64:  DefaultFmtInt64Assignment,
			types.Int:    DefaultFmtInt64Assignment,
			types.Uint8:  DefaultFmtInt64Assignment,
			types.Uint64: DefaultFmtInt64Assignment,
		},
	}
	for _, f := range opts {
		f(b)
	}
	return b
}

type Basic struct {
	Templates map[types.BasicKind]string
}

func (s *Basic) Print(a, b *types.Basic, aFieldPath, bFieldPath string, levelNum int) string {
	if a.Kind() != b.Kind() {
		return fmt.Sprintf("not same basic kind in %s")
	}
	result, ok := s.Templates[a.Kind()]
	if !ok {
		return fmt.Sprintf("unknown basic kind: %d\n", a.Kind())
	}
	result = strings.ReplaceAll(result, "$a", aFieldPath)
	result = strings.ReplaceAll(result, "$b", bFieldPath)
	return result + "\n"
}