package traverser

import (
	"bytes"
	"fmt"
	"github.com/muvaf/typewriter/pkg/imports"
	"go/types"
	"text/template"

	"github.com/pkg/errors"
)

func NewMap(im imports.Map) *Map {
	return &Map{
		Imports: im,
	}
}

// NOTE(muvaf): Statement should not have any tabs because it is multi-line and
// each line has their own tab space. Hence it only helps the first line, which
// is empty anyway.

const DefaultMapTmpl = `
if len({{ .AFieldPath }}) != 0 {
  {{ .BFieldPath }} = make({{ .TypeB }}, len({{ .AFieldPath }}))
  for {{ .Key }} := range {{ .AFieldPath }} {
{{ .Statements }}
  }
}`

type DefaultMapTmplInput struct {
	AFieldPath string
	TypeA      string
	BFieldPath string
	TypeB      string
	Key      string
	Value      string
	Statements  string
}

type Map struct {
	Imports imports.Map
	Recursive TypeTraverser
}

func (s *Map) SetTypeTraverser(p TypeTraverser) {
	s.Recursive = p
}

func (s *Map) Print(a, b *types.Map, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	key := fmt.Sprintf("k%d", levelNum)
	statements, err := s.Recursive.Print(a.Elem(), b.Elem(), fmt.Sprintf("%s[%s]", aFieldPath, key), fmt.Sprintf("%s[%s]", bFieldPath, key), levelNum+1)
	if err != nil {
		return "", errors.Wrap(err, "cannot recursively traverse element type of slice")
	}
	i := DefaultMapTmplInput{
		AFieldPath: aFieldPath,
		TypeA:      s.Imports.UseType(a.String()),
		BFieldPath: bFieldPath,
		TypeB:      s.Imports.UseType(b.String()),
		Key:      key,
		Statements:  statements,
	}
	t, err := template.New("func").Parse(DefaultMapTmpl)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
