// Copyright 2022 Muvaffak Onus
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

package packages

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/muvaf/typewriter/pkg/test"
)

type importsModifier func(i *Imports)

func importsWithImportsMap(v map[string]string) importsModifier {
	return func(i *Imports) {
		i.Imports = v
	}
}

func imports(cm ...importsModifier) *Imports {
	c := &Imports{
		Imports: map[string]string{},
	}
	for _, m := range cm {
		m(c)
	}
	return c
}

func TestParseTypeDec(t *testing.T) {
	type args struct {
		s string
	}
	type want struct {
		pkgName string
		field   string
	}
	cases := map[string]struct {
		args
		want
	}{
		"String": {
			args: args{s: "string"},
			want: want{pkgName: "", field: "string"},
		},
		"StringPointer": {
			args: args{s: "*string"},
			want: want{pkgName: "", field: "*string"},
		},
		"Bool": {
			args: args{s: "bool"},
			want: want{pkgName: "", field: "bool"},
		},
		"Int": {
			args: args{s: "int"},
			want: want{pkgName: "", field: "int"},
		},
		"Slice": {
			args: args{s: "[]v1alpha1.ExampleStruct"},
			want: want{pkgName: "v1alpha1", field: "[]%s.ExampleStruct"},
		},
		"SlicePointer": {
			args: args{s: "[]*v1alpha1.ExampleStruct"},
			want: want{pkgName: "v1alpha1", field: "[]*%s.ExampleStruct"},
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			pkgName, field := parseTypeDec(tc.s)
			if diff := cmp.Diff(tc.want.pkgName, pkgName, test.EquateErrors()); diff != "" {
				t.Errorf("parseTypeDec(...) pkgName = %v, want %v", pkgName, tc.want.pkgName)
			}
			if diff := cmp.Diff(tc.want.field, field); diff != "" {
				t.Errorf("parseTypeDec(...) field = %v, want %v", field, tc.want.field)
			}
		})
	}
}

func TestImports_UseType(t *testing.T) {
	type args struct {
		m  *Imports
		in string
	}
	type want struct {
		m        *Imports
		typeName string
	}
	cases := map[string]struct {
		args
		want
	}{
		"String": {
			args: args{
				m:  imports(),
				in: "string",
			},
			want: want{
				m:        imports(),
				typeName: "string",
			},
		},
		"MapStringToStruct": {
			args: args{
				m:  imports(),
				in: "map[string]github.com/org/repo/v1alpha1.ExampleStruct",
			},
			want: want{
				m: imports(
					importsWithImportsMap(map[string]string{
						"github.com/org/repo/v1alpha1": "v1alpha1",
					}),
				),
				typeName: "map[string]v1alpha1.ExampleStruct",
			},
		},
		"MapStructToStruct": {
			args: args{
				m: imports(
					importsWithImportsMap(map[string]string{
						"github.com/example/ex/pkg": "exPkg",
					}),
				),
				in: "map[github.com/org/repo/v1alpha1.ExampleStruct]github.com/org/repo/v1beta1.ExampleStruct1",
			},
			want: want{
				m: imports(
					importsWithImportsMap(map[string]string{
						"github.com/example/ex/pkg":    "exPkg",
						"github.com/org/repo/v1alpha1": "v1alpha1",
						"github.com/org/repo/v1beta1":  "v1beta1",
					}),
				),
				typeName: "map[v1alpha1.ExampleStruct]v1beta1.ExampleStruct1",
			},
		},
	}
	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			gotTypeName := tc.args.m.UseType(tc.in)
			if diff := cmp.Diff(tc.want.typeName, gotTypeName, test.EquateErrors()); diff != "" {
				t.Errorf("useType(...) pkgName = %v, want %v", gotTypeName, tc.want.typeName)
			}
			if diff := cmp.Diff(tc.args.m, tc.want.m, test.EquateErrors()); diff != "" {
				t.Errorf("useType(...) imports = %v, want %v", gotTypeName, tc.want.typeName)
			}
		})
	}
}
