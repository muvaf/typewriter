package scanner

import (
	"bytes"
	"go/types"
	"strings"
	"text/template"

	"github.com/muvaf/typewriter/pkg/packages"

	"github.com/pkg/errors"
)

func NewTypePrinter(originPackagePath string, rootType *types.Named, commentMarkers string, im *packages.Imports, targetScope *types.Scope) *TypePrinter {
	return &TypePrinter{
		OriginPackagePath: originPackagePath,
		RootType:          rootType,
		CommentMarkers:    commentMarkers,
		Imports:           im,
		TypeMap:           map[string]*types.Named{},
		TargetScope:       targetScope,
	}
}

type TypePrinter struct {
	OriginPackagePath string
	RootType          *types.Named
	CommentMarkers    string
	Imports           *packages.Imports
	TypeMap           map[string]*types.Named
	TargetScope       *types.Scope
}

func (tp *TypePrinter) Parse() {
	tp.load(tp.RootType)
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
{{- .CommentMarkers }}
type {{ .Name }} struct {
{{ .Fields }}
}`
	FieldTmpl = `
{{ .Comment }}
{{ .Name }} {{ .Type }} {{ .Tags }}`
)

type TypeTmplInput struct {
	Name           string
	Fields         string
	Comment        string
	CommentMarkers string
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
		if name == tp.RootType.Obj().Name() {
			ti.Name = targetName
			ti.CommentMarkers = tp.CommentMarkers
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
		ti.Fields = strings.ReplaceAll(ti.Fields, "\n\n", "\n")
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
