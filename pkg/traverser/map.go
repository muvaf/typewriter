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
	Key        string
	Value      string
	Statements string
}

func NewMap(im *packages.Imports) *Map {
	return &Map{
		Template: DefaultMapTmpl,
		Imports:  im,
	}
}

type Map struct {
	Template string
	Imports  *packages.Imports
	Generic  GenericTraverser
}

func (m *Map) SetTemplate(t string) {
	m.Template = t
}

func (m *Map) SetGenericTraverser(p GenericTraverser) {
	m.Generic = p
}

func (m *Map) Print(a, b *types.Map, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	key := fmt.Sprintf("k%d", levelNum)
	statements, err := m.Generic.Print(a.Elem(), b.Elem(), fmt.Sprintf("%s[%s]", aFieldPath, key), fmt.Sprintf("%s[%s]", bFieldPath, key), levelNum+1)
	if err != nil {
		return "", errors.Wrap(err, "cannot recursively traverse element type of slice")
	}
	i := DefaultMapTmplInput{
		AFieldPath: aFieldPath,
		TypeA:      m.Imports.UseType(a.String()),
		BFieldPath: bFieldPath,
		TypeB:      m.Imports.UseType(b.String()),
		Key:        key,
		Statements: statements,
	}
	t, err := template.New("func").Parse(m.Template)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
