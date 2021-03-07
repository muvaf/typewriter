package typewriter

import (
	"fmt"
	"go/types"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/muvaf/typewriter/pkg/test"
)

const simpleTypes = `
package test

type A struct {
	One string
	Two *string
	Three int64
}

type B struct {
	One string
	Two *string
	Three int64
	Four *int64
}
`

func TestPrint(t *testing.T) {
	s := test.ParseString(simpleTypes)
	aType := s.Lookup("A").Type()
	bType := s.Lookup("B").Type()
	type args struct {
		a     *types.Basic
		b     *types.Basic
		aPath string
		bPath string
	}
	type want struct {
		out string
		err error
	}
	cases := map[string]struct {
		args
		want
	}{
		"Success": {
			args: args{
				a:     aType.(*types.Named).Underlying().(*types.Struct).Field(0).Type().(*types.Basic),
				b:     bType.(*types.Named).Underlying().(*types.Struct).Field(0).Type().(*types.Basic),
				aPath: "a",
				bPath: "b",
			},
			want: want{
				out: "\na = b",
			},
		},
		"ErrTypeMismatch": {
			args: args{
				a:     aType.(*types.Named).Underlying().(*types.Struct).Field(0).Type().(*types.Basic),
				b:     bType.(*types.Named).Underlying().(*types.Struct).Field(2).Type().(*types.Basic),
				aPath: "a",
				bPath: "b",
			},
			want: want{
				out: "",
				err: fmt.Errorf(errFmtNotSameKind, "string", "int64"),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			b := NewBasic(WithTmpl(types.String, AssignmentTmpl))
			result, err := b.Print(tc.args.a, tc.args.b, tc.args.aPath, tc.args.bPath)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("add: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.out, result); diff != "" {
				t.Errorf("add: -want, +got:\n%s", diff)
			}
		})
	}
}
