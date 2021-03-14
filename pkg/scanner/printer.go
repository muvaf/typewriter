package scanner

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
	"text/template"

	"github.com/muvaf/typewriter/pkg/imports"

	"github.com/pkg/errors"
)

func NewTypePrinter(originPackagePath string, rootType *Named, im *imports.Map, targetScope *types.Scope) *TypePrinter {
	return &TypePrinter{
		OriginPackagePath: originPackagePath,
		RootType:          rootType,
		Imports:           im,
		TypeMap:           map[string]*types.Named{},
		TargetScope:       targetScope,
	}
}

type TypePrinter struct {
	OriginPackagePath string
	RootType          *Named
	Imports           *imports.Map
	TypeMap           map[string]*types.Named
	TargetScope       *types.Scope
}

func (tp *TypePrinter) Parse() {
	tp.load(tp.RootType.Named)
}

func (tp *TypePrinter) load(t *types.Named) {
	s, ok := t.Underlying().(*types.Struct)
	if !ok {
		// might be function
		return
	}
	// todo: naming collisions? is it possible this function runs with multiple
	// packages?
	tp.TypeMap[t.Obj().Name()] = t
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

// We could have a for look in TypeTmpl but I don't want to run any function
// in Go templates as it's hard to control, easy to make mistakes and too rigid
// for exception cases.

const (
	TypeTmpl = `
{{ .Comment }}
{{- .CommentTags }}
type {{ .Name }} struct {
{{ .Fields }}
}`
	FieldTmpl = `
{{ .Comment }}
{{ .Name }} {{ .Type }} {{ .Tags }}`
)

type TypeTmplInput struct {
	Name        string
	Fields      string
	Comment     string
	CommentTags string
}

type FieldTmplInput struct {
	Name    string
	Type    string
	Tags    string
	Comment string
}

func (tp *TypePrinter) Print(targetName string) (string, error) {
	out := ""
	for name, n := range tp.TypeMap {
		ti := &TypeTmplInput{
			Name: name,
		}
		if name == tp.RootType.Named.Obj().Name() {
			ti.Name = targetName
			commentTags := ""
			for _, tag := range tp.RootType.CommentTags {
				commentTags = fmt.Sprintf("%s\n%s", commentTags, tag)
			}
			ti.CommentTags = commentTags
		}
		// If the type already exists in the package, we assume it's the same
		// as the one we use here.
		if tp.TargetScope.Lookup(ti.Name) != nil {
			continue
		}
		s := n.Underlying().(*types.Struct)
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
		tp.TargetScope.Insert(n.Obj())
		out += result.String()
	}
	return out, nil
}
