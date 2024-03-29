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
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

const DirectProducerTmpl = `
// {{ .FunctionName }} returns a new {{ .BTypeName }} with the information from
// given {{ .ATypeName }}.
func {{ .FunctionName }}(a {{ .ATypeName }}) {{ .BTypeName }} {
  b := {{ .BTypeNewStatement }}
{{ .Statements }}
  return b
}`

func WithTemplate(t string) PrinterOption {
	return func(p *Printer) {
		p.Template = t
	}
}

type PrinterOption func(p *Printer)

func NewPrinter(im *packages.Imports, tr GenericTraverser, opts ...PrinterOption) *Printer {
	f := &Printer{
		Imports:   im,
		Traverser: tr,
		Template:  DirectProducerTmpl,
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

type Printer struct {
	Imports   *packages.Imports
	Traverser GenericTraverser
	Template  string
}

func (p *Printer) Print(name string, a, b types.Type, extraInput map[string]interface{}) (string, error) {
	content, err := p.Traverser.Print(a, b, "a", "b", 0)
	if err != nil {
		return "", errors.Wrap(err, "cannot traverse")
	}
	var an *types.Named
	aNamePrefix := ""
	var bn *types.Named
	bNamePrefix := ""
	switch at := a.(type) {
	case *types.Pointer:
		an = at.Underlying().(*types.Named)
		aNamePrefix = "*"
	default:
		an = a.(*types.Named)
	}
	switch bt := b.(type) {
	case *types.Pointer:
		bn = bt.Underlying().(*types.Named)
		bNamePrefix = "*"
	default:
		bn = b.(*types.Named)
	}
	aTypeDec := p.Imports.UseType(an.String())
	aTypeName := fmt.Sprintf("%s%s", aNamePrefix, aTypeDec)
	aNewStatement := fmt.Sprintf("%s{}", aTypeName)
	if aNamePrefix == "*" {
		aNewStatement = fmt.Sprintf("&%s", aNewStatement)
	}
	bTypeDec := p.Imports.UseType(bn.String())
	bTypeName := fmt.Sprintf("%s%s", bNamePrefix, bTypeDec)
	bNewStatement := fmt.Sprintf("%s{}", bTypeName)
	if bNamePrefix == "*" {
		bNewStatement = fmt.Sprintf("&%s", bNewStatement)
	}
	ts := map[string]interface{}{
		"FunctionName":      name,
		"ATypeName":         aTypeName,
		"ATypeNewStatement": aNewStatement,
		"BTypeName":         bTypeName,
		"BTypeNewStatement": bNewStatement,
		"Statements":        content,
	}
	for k, v := range extraInput {
		ts[k] = v
	}
	t, err := template.New("func").Parse(p.Template)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, ts)
	return strings.ReplaceAll(result.String(), "\n\n", "\n"), errors.Wrap(err, "cannot execute template")
}
