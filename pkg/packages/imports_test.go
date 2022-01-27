package packages

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/muvaf/typewriter/pkg/test"
)

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
		"BuiltInMap": {
			args: args{s: "map[string]string"},
			want: want{pkgName: "", field: "map[string]string"},
		},
		"MapToStruct": {
			args: args{s: "map[string]v1alpha1.ExampleStruct"},
			want: want{pkgName: "v1alpha1", field: "map[string]%s.ExampleStruct"},
		},
		"MapToStructPointer": {
			args: args{s: "map[string]*v1alpha1.ExampleStruct"},
			want: want{pkgName: "v1alpha1", field: "map[string]*%s.ExampleStruct"},
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			pkgName, field := parseTypeDec(tc.s)
			if diff := cmp.Diff(tc.want.pkgName, pkgName, test.EquateErrors()); diff != "" {
				t.Errorf("generateTypeName(...) pkgName = %v, want %v", pkgName, tc.want.pkgName)
			}
			if diff := cmp.Diff(tc.want.field, field); diff != "" {
				t.Errorf("generateTypeName(...) field = %v, want %v", field, tc.want.field)
			}
		})
	}
}
