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

package traverser

import (
	"bytes"
	"fmt"
	"go/types"
	"text/template"

	"github.com/muvaf/typewriter/pkg/packages"

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
	Statements string
}

func NewSlice(im *packages.Imports) *Slice {
	return &Slice{
		Imports:  im,
		Template: DefaultSliceTmpl,
	}
}

type Slice struct {
	Template string
	Imports  *packages.Imports
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
		Statements: statements,
	}
	t, err := template.New("func").Parse(s.Template)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
