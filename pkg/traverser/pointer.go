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
	"go/types"
	"strings"
	"text/template"

	"github.com/muvaf/typewriter/pkg/packages"

	"github.com/pkg/errors"
)

// NOTE(muvaf): Statement should not have any tabs because it is multi-line and
// each line has their own tab space. Hence it only helps the first line, which
// is empty anyway.

const DefaultPointerTmpl = `
if {{ .AFieldPath }} != nil {
  {{ .BFieldPath }} = new({{ .NonPointerTypeB }})
{{ .Statements }}
}`

type PointerTmplInput struct {
	AFieldPath      string
	TypeA           string
	NonPointerTypeA string
	BFieldPath      string
	TypeB           string
	NonPointerTypeB string
	Statements      string
}

func NewPointer(im *packages.Imports) *Pointer {
	return &Pointer{
		Template: DefaultPointerTmpl,
		Imports:  im,
	}
}

type Pointer struct {
	Template string
	Imports  *packages.Imports
	Generic  GenericTraverser
}

func (p *Pointer) SetTemplate(t string) {
	p.Template = t
}

func (p *Pointer) SetGenericTraverser(tt GenericTraverser) {
	p.Generic = tt
}

func (p *Pointer) Print(a, b *types.Pointer, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	statements, err := p.Generic.Print(a.Elem(), b.Elem(), aFieldPath, bFieldPath, levelNum)
	if err != nil {
		return "", errors.Wrap(err, "cannot recursively traverse element type of pointer")
	}
	i := PointerTmplInput{
		AFieldPath:      aFieldPath,
		TypeA:           p.Imports.UseType(a.String()),
		NonPointerTypeA: strings.TrimPrefix(p.Imports.UseType(a.String()), "*"),
		BFieldPath:      bFieldPath,
		TypeB:           p.Imports.UseType(b.String()),
		NonPointerTypeB: strings.TrimPrefix(p.Imports.UseType(b.String()), "*"),
		Statements:      statements,
	}
	t, err := template.New("func").Parse(p.Template)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
