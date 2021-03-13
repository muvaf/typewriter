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

const DefaultMapTmpl = `
if len({{ .AFieldPath }}) != 0 {
  {{ .BFieldPath }} = make({{ .TypeB }}, len({{ .AFieldPath }}))
  for {{ .Key }} := range {{ .AFieldPath }} {
    {{ .Statement }}
  }
}`

type DefaultMapTmplInput struct {
	AFieldPath string
	TypeA      string
	BFieldPath string
	TypeB      string
	Key      string
	Value      string
	Statement  string
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
	statement, err := s.Recursive.Print(a.Elem(), b.Elem(), fmt.Sprintf("%s[%s]", aFieldPath, key), fmt.Sprintf("%s[%s]", bFieldPath, key), levelNum+1)
	if err != nil {
		return "", errors.Wrap(err, "cannot recursively traverse element type of slice")
	}
	i := DefaultMapTmplInput{
		AFieldPath: aFieldPath,
		TypeA:      s.Imports.UseType(a.String()),
		BFieldPath: bFieldPath,
		TypeB:      s.Imports.UseType(b.String()),
		Key:      key,
		Statement:  statement,
	}
	t, err := template.New("func").Parse(DefaultMapTmpl)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
