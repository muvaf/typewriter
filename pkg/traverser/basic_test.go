// Copyright 2021 Muvaffak Onus
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package traverser

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
				out: "\nb = a",
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
			b := NewBasic()
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
