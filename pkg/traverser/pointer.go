package traverser

import (
	"bytes"
	"github.com/muvaf/typewriter/pkg/imports"
	"go/types"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

func NewPointer(im imports.Map) *Pointer {
	return &Pointer{
		Imports: im,
	}
}

// NOTE(muvaf): Statement should not have any tabs because it is multi-line and
// each line has their own tab space. Hence it only helps the first line, which
// is empty anyway.

const DefaultPointerTmpl = `
if {{ .AFieldPath }} != nil {
  {{ .BFieldPath }} = new({{ .NonPointerTypeB }})
{{ .Statements }}
}`

type DefaultPointerTmplInput struct {
	AFieldPath string
	TypeA      string
	NonPointerTypeA      string
	BFieldPath string
	TypeB      string
	NonPointerTypeB      string
	Statements  string
}

type Pointer struct {
	Imports imports.Map
	Recursive TypeTraverser
}

func (s *Pointer) SetTypeTraverser(p TypeTraverser) {
	s.Recursive = p
}

func (s *Pointer) Print(a, b *types.Pointer, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	statements, err := s.Recursive.Print(a.Elem(), b.Elem(), aFieldPath, bFieldPath, levelNum)
	if err != nil {
		return "", errors.Wrap(err, "cannot recursively traverse element type of pointer")
	}
	i := DefaultPointerTmplInput{
		AFieldPath: aFieldPath,
		TypeA:      s.Imports.UseType(a.String()),
		NonPointerTypeA:      strings.TrimPrefix(s.Imports.UseType(a.String()), "*"),
		BFieldPath: bFieldPath,
		TypeB:      s.Imports.UseType(b.String()),
		NonPointerTypeB:      strings.TrimPrefix(s.Imports.UseType(b.String()), "*"),
		Statements:  statements,
	}
	t, err := template.New("func").Parse(DefaultPointerTmpl)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
