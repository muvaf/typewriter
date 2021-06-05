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

package test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"

	"github.com/google/go-cmp/cmp"
)

func EquateErrors() cmp.Option {
	return cmp.Comparer(func(a, b error) bool {
		if a == nil || b == nil {
			return a == nil && b == nil
		}

		av := reflect.ValueOf(a)
		bv := reflect.ValueOf(b)
		if av.Type() != bv.Type() {
			return false
		}

		return a.Error() == b.Error()
	})
}

func ParseString(s string) *types.Scope {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "simple.go", s, 0)
	if err != nil {
		panic(err)
	}
	cfg := types.Config{Importer: importer.Default()}
	pkg, err := cfg.Check("simple.go", fset, []*ast.File{f}, nil)
	if err != nil {
		panic(err)
	}
	return pkg.Scope()
}
