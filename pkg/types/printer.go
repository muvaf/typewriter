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

package types

import (
	"bytes"
	"fmt"
	"go/types"
	"sort"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

// We could have a for loop in StructTypeTmpl but I don't want to run any function
// in Go templates as it's hard to control, easy to make mistakes and too rigid
// for exception cases. Though we could test whether calling different templating
// functions independently cause performance problems. Maybe call them in parallel?

const (
	StructTypeTmpl = `

{{ .Comment }}
type {{ .Name }} struct {
{{ .Fields }}
}`
	FieldTmpl    = "\n\n\n{{ .Comment }}\n{{ .Name }} {{ .Type }} `{{ .Tag }}`"
	EnumTypeTmpl = `

{{ .Comment }}
type {{ .Name }} {{ .UnderlyingType }}`
)

type StructTypeTmplInput struct {
	Name    string
	Fields  string
	Comment string
}

type FieldTmplInput struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

type EnumTypeTmplInput struct {
	Name           string
	UnderlyingType string
	Comment        string
}

func WithComments(c Comments) PrinterOption {
	return func(p *Printer) {
		p.Comments = c
	}
}

type PrinterOption func(*Printer)

func NewPrinter(im *packages.Imports, targetScope *types.Scope, opts ...PrinterOption) *Printer {
	p := &Printer{
		Imports:     im,
		TargetScope: targetScope,
		Comments:    Comments{},
	}

	for _, f := range opts {
		f(p)
	}
	return p
}

type Printer struct {
	Imports     *packages.Imports
	TargetScope *types.Scope
	Comments    Comments
}

func (tp *Printer) Print(typeList []*types.Named) (string, error) {
	out := ""
	sort.SliceStable(typeList, func(i, j int) bool {
		return typeList[i].Obj().Name() < typeList[j].Obj().Name()
	})
	for _, n := range typeList {
		// If the type already exists in the package, we assume it's the same
		// as the one we use here.
		if tp.TargetScope.Lookup(n.Obj().Name()) != nil {
			continue
		}
		switch o := n.Underlying().(type) {
		case *types.Struct:
			result, err := tp.printStructType(n.Obj(), o)
			if err != nil {
				return "", errors.Wrapf(err, "cannot print struct type %s", n.Obj().Name())
			}
			out += result

		case *types.Basic:
			result, err := tp.printEnumType(n.Obj(), o)
			if err != nil {
				return "", errors.Wrapf(err, "cannot print struct type %s", n.Obj().Name())
			}
			out += result
		default:
			fmt.Printf("underlying of the type is neither Struct nor Basic, skipping %s\n", n.Obj().Name())
			continue
		}
		tp.TargetScope.Insert(n.Obj())
	}
	return out, nil
}

// printEnumType assumes that the underlying type is a basic type, which may not
// be the case all the time.
// TODO(muvaf): Think about how to handle `type MyEnum MyOtherType`
func (tp *Printer) printEnumType(name *types.TypeName, b *types.Basic) (string, error) {
	ei := &EnumTypeTmplInput{
		Name:           name.Name(),
		UnderlyingType: b.Name(),
		Comment:        tp.Comments[QualifiedTypePath(name)],
	}
	t, err := template.New("enum").Parse(EnumTypeTmpl)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	if err = t.Execute(result, ei); err != nil {
		return "", errors.Wrap(err, "cannot execute templating")
	}
	return result.String(), nil
}

func (tp *Printer) printStructType(name *types.TypeName, s *types.Struct) (string, error) {
	ti := &StructTypeTmplInput{
		Name:    name.Name(),
		Comment: tp.Comments[QualifiedTypePath(name)],
	}
	// Field order we get here is not stable but tag & field indexes are coupled.
	tagMap := make(map[*types.Var]string, s.NumFields())
	for i := 0; i < s.NumFields(); i++ {
		tagMap[s.Field(i)] = s.Tag(i)
	}
	fields := make([]*types.Var, len(tagMap))
	i := 0
	for f := range tagMap {
		fields[i] = f
		i++
	}
	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].Name() < fields[j].Name()
	})
	for _, field := range fields {
		fi := &FieldTmplInput{
			Name:    field.Name(),
			Type:    tp.Imports.UseType(field.Type().String()),
			Tag:     tagMap[field],
			Comment: tp.Comments[QualifiedFieldPath(name, field.Name())],
		}
		t, err := template.New("func").Parse(FieldTmpl)
		if err != nil {
			return "", errors.Wrap(err, "cannot parse template")
		}
		result := &bytes.Buffer{}
		if err = t.Execute(result, fi); err != nil {
			return "", errors.Wrap(err, "cannot execute templating")
		}
		ti.Fields += result.String()
	}
	ti.Fields = strings.ReplaceAll(ti.Fields, "\n\n", "\n")
	t, err := template.New("func").Parse(StructTypeTmpl)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	if err = t.Execute(result, ti); err != nil {
		return "", errors.Wrap(err, "cannot execute templating")
	}
	return result.String(), nil
}
