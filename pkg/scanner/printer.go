package scanner

import (
	"bytes"
	"go/types"
	"strings"
	"text/template"

	"github.com/muvaf/typewriter/pkg/imports"

	"github.com/pkg/errors"
)

func NewTypePrinter(p string, im *imports.Map) *TypePrinter {
	return &TypePrinter{
		OriginPackagePath: p,
		Imports:           im,
		TypeMap:           map[string]*types.Struct{},
	}
}

type TypePrinter struct {
	OriginPackagePath string
	Imports           *imports.Map
	TypeMap           map[string]*types.Struct
}

func (tp *TypePrinter) Load(t *types.Named) {
	s, ok := t.Underlying().(*types.Struct)
	if !ok {
		// might be function
		return
	}
	// todo: naming collisions? is it possible this function runs with multiple
	// packages?
	tp.TypeMap[t.Obj().Name()] = s
	for i := 0; i < s.NumFields(); i++ {
		ft := s.Field(i).Type()
		switch u := ft.(type) {
		case *types.Pointer:
			n, ok := u.Elem().(*types.Named)
			if !ok {
				continue
			}
			tp.Load(n)
		case *types.Slice:
			switch n := u.Elem().(type) {
			case *types.Named:
				tp.Load(n)
			case *types.Pointer:
				pn, ok := n.Elem().(*types.Named)
				if !ok {
					continue
				}
				tp.Load(pn)
			}
		case *types.Named:
			tp.Load(u)
		}
	}
}

// We could have a for look in TypeTmpl but I don't want to run any function
// in Go templates as it's hard to control, easy to make mistakes and too rigid
// for exception cases.

const (
	TypeTmpl = `
{{ .Comment }}
type {{ .Name }} struct {
{{ .Fields }}
}`
	FieldTmpl = `
{{ .Comment }}
{{ .Name }} {{ .Type }} {{ .Tags }}`
)

type TypeTmplInput struct {
	Name    string
	Fields  string
	Comment string
}

type FieldTmplInput struct {
	Name    string
	Type    string
	Tags    string
	Comment string
}

func (tp *TypePrinter) Print() (string, error) {
	out := ""
	for name, s := range tp.TypeMap {
		ti := &TypeTmplInput{
			Name: name,
		}
		for i := 0; i < s.NumFields(); i++ {
			f := s.Field(i)
			// The structs in the remote package are known to be copied, so the
			// types should reference the local copies.
			remoteType := strings.ReplaceAll(f.Type().String(), tp.OriginPackagePath, tp.Imports.Package)
			fi := &FieldTmplInput{
				Name: f.Name(),
				Type: tp.Imports.UseType(remoteType),
				//Tags: f.
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
		t, err := template.New("func").Parse(TypeTmpl)
		if err != nil {
			return "", errors.Wrap(err, "cannot parse template")
		}
		result := &bytes.Buffer{}
		if err = t.Execute(result, ti); err != nil {
			return "", errors.Wrap(err, "cannot execute templating")
		}
		out += result.String()
	}
	return out, nil
}
