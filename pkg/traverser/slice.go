package traverser

import (
	"bytes"
	"fmt"
	"github.com/muvaf/typewriter/pkg/imports"
	"go/types"
	"text/template"

	"github.com/pkg/errors"
)

const DefaultSliceTmpl = `
if len({{ .AFieldPath }}) != 0 {
  {{ .BFieldPath }} = make({{ .TypeB }}, len({{ .AFieldPath }}))
  for {{ .Index }} := range {{ .AFieldPath }} {
{{ .Statements }}
  }
}`

type SliceTmplInput struct {
	AFieldPath string
	TypeA      string
	BFieldPath string
	TypeB      string
	Index      string
	Statements  string
}

func NewSlice(im imports.Map) *Slice {
	return &Slice{
		Imports: im,
		Template: DefaultSliceTmpl,
	}
}

type Slice struct {
	Template string
	Imports  imports.Map
	Generic  GenericTraverser
}

func (s *Slice) SetTemplate(t string) {
	s.Template = t
}

func (s *Slice) SetGenericTraverser(p GenericTraverser) {
	s.Generic = p
}

func (s *Slice) Print(a, b *types.Slice, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	index := fmt.Sprintf("v%d", levelNum)
	statements, err := s.Generic.Print(a.Elem(), b.Elem(), fmt.Sprintf("%s[%s]", aFieldPath, index), fmt.Sprintf("%s[%s]", bFieldPath, index), levelNum+1)
	if err != nil {
		return "", errors.Wrap(err, "cannot recursively traverse element type of slice")
	}
	i := SliceTmplInput{
		AFieldPath: aFieldPath,
		TypeA:      s.Imports.UseType(a.String()),
		BFieldPath: bFieldPath,
		TypeB:      s.Imports.UseType(b.String()),
		Index:      index,
		Statements:  statements,
	}
	t, err := template.New("func").Parse(s.Template)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
