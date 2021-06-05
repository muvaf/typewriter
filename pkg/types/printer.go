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
	"go/types"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

func NewTypePrinter(im *packages.Imports, targetScope *types.Scope) *Printer {
	return &Printer{
		Imports:     im,
		TypeMap:     map[*types.TypeName]types.Type{},
		TargetScope: targetScope,
	}
}

type Printer struct {
	Imports     *packages.Imports
	TypeMap     map[*types.TypeName]types.Type
	TargetScope *types.Scope
}

func (tp *Printer) load(t *types.Named) {
	if t.Underlying() != nil {
		tp.TypeMap[t.Obj()] = t.Underlying()
	}
	// todo: naming collisions? is it possible this function runs with multiple
	// packages?
	s, ok := t.Underlying().(*types.Struct)
	if !ok {
		return
	}
	for i := 0; i < s.NumFields(); i++ {
		ft := s.Field(i).Type()
		switch u := ft.(type) {
		case *types.Pointer:
			n, ok := u.Elem().(*types.Named)
			if !ok {
				continue
			}
			tp.load(n)
		case *types.Slice:
			switch n := u.Elem().(type) {
			case *types.Named:
				tp.load(n)
			case *types.Pointer:
				pn, ok := n.Elem().(*types.Named)
				if !ok {
					continue
				}
				tp.load(pn)
			}
		case *types.Named:
			tp.load(u)
		}
	}
}

// We could have a for loop in StructTypeTmpl but I don't want to run any function
// in Go templates as it's hard to control, easy to make mistakes and too rigid
// for exception cases. Though we could test whether calling different templating
// functions independently cause performance problems. Maybe call them in parallel?

const (
	StructTypeTmpl = `
{{ .Comment }}
{{- .CommentMarkers }}
type {{ .Name }} struct {
{{ .Fields }}
}`
	FieldTmpl    = "\n{{ .Comment }}\n{{ .CommentMarkers }}\n{{ .Name }} {{ .Type }} `{{ .Tag }}`"
	EnumTypeTmpl = `
{{ .Comment }}
{{- .CommentMarkers }}
type {{ .Name }} {{ .UnderlyingType }}`
)

type StructTypeTmplInput struct {
	Name           string
	Fields         string
	Comment        string
	CommentMarkers string
}

type FieldTmplInput struct {
	Name           string
	Type           string
	Tag            string
	Comment        string
	CommentMarkers string
}

type EnumTypeTmplInput struct {
	Name           string
	UnderlyingType string
	Comment        string
	CommentMarkers string
}

func (tp *Printer) Print(rootType *types.Named, commentMarkers string) (string, error) {
	tp.load(rootType)
	out := ""
	for name, n := range tp.TypeMap {
		// If the type already exists in the package, we assume it's the same
		// as the one we use here.
		if tp.TargetScope.Lookup(name.Name()) != nil {
			continue
		}
		markers := ""
		if name.Name() == rootType.Obj().Name() {
			markers = commentMarkers
		}
		switch o := n.Underlying().(type) {
		case *types.Struct:
			result, err := tp.printStructType(name, o, markers)
			if err != nil {
				return "", errors.Wrapf(err, "cannot print struct type %s", name.Name())
			}
			out += result
		case *types.Basic:
			result, err := tp.printEnumType(name, o, markers)
			if err != nil {
				return "", errors.Wrapf(err, "cannot print struct type %s", name.Name())
			}
			out += result
		}
		tp.TargetScope.Insert(name)
	}
	return out, nil
}

// printEnumType assumes that the underlying type is a basic type, which may not
// be the case all the time.
// TODO(muvaf): Think about how to handle `type MyEnum MyOtherType`
func (tp *Printer) printEnumType(name *types.TypeName, b *types.Basic, commentMarkers string) (string, error) {
	ei := &EnumTypeTmplInput{
		Name:           name.Name(),
		CommentMarkers: commentMarkers,
		UnderlyingType: b.Name(),
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

func (tp *Printer) printStructType(name *types.TypeName, s *types.Struct, commentMarkers string) (string, error) {
	ti := &StructTypeTmplInput{
		Name:           name.Name(),
		CommentMarkers: commentMarkers,
	}
	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		tag := s.Tag(i)
		// The structs in the remote package are known to be copied, so the
		// types should reference the local copies.
		remoteType := field.Type().String()
		var tnamed *types.Named
		switch o := field.Type().(type) {
		case *types.Pointer:
			tn, ok := o.Elem().(*types.Named)
			if ok {
				tnamed = tn
			}
		case *types.Slice:
			tn, ok := o.Elem().(*types.Named)
			if ok {
				tnamed = tn
			}
		case *types.Map:
			tn, ok := o.Elem().(*types.Named)
			if ok {
				tnamed = tn
			}
		case *types.Named:
			tnamed = o
		}
		if tnamed != nil {
			remoteType = strings.ReplaceAll(field.Type().String(), tnamed.Obj().Pkg().Path(), tp.Imports.Package)
		}
		fi := &FieldTmplInput{
			Name: field.Name(),
			Type: tp.Imports.UseType(remoteType),
			Tag:  tag,
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
