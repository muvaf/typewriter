package typewriter

import (
	"bytes"
	"fmt"
	"go/types"
	"text/template"

	"github.com/pkg/errors"
)

func NewSlice() *Slice {
	return &Slice{}
}

const DefaultSliceTmpl = `
if len({{ .AFieldPath }}) != 0 {
  {{ .BFieldPath }} := make({{ .TypeB }}, len({{ .AFieldPath }}))
  for {{ .Index }} := range {{ .AFieldPath }} {
    {{ .Statement }}
  }
}`

type DefaultSliceTmplInput struct {
	AFieldPath string
	TypeA      string
	BFieldPath string
	TypeB      string
	Index      string
	Statement  string
}

type Slice struct {
	Recursive TypeTraverser
}

func (s *Slice) SetTypeTraverser(p TypeTraverser) {
	s.Recursive = p
}

func (s *Slice) Print(a, b *types.Slice, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	index := fmt.Sprintf("v%d", levelNum)
	statement, err := s.Recursive.Print(a.Elem(), b.Elem(), fmt.Sprintf("%s[%s]", aFieldPath, index), fmt.Sprintf("%s[%s]", bFieldPath, index), levelNum+1)
	if err != nil {
		return "", errors.Wrap(err, "cannot recursively traverse element type of slice")
	}
	i := DefaultSliceTmplInput{
		AFieldPath: aFieldPath,
		TypeA:      a.String(),
		BFieldPath: bFieldPath,
		TypeB:      b.String(),
		Index:      index,
		Statement:  statement,
	}
	t, err := template.New("func").Parse(DefaultSliceTmpl)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
