package wrapper

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
	"text/template"

	"github.com/muvaf/typewriter/pkg/traverser"

	"github.com/muvaf/typewriter/pkg/packages"

	"github.com/pkg/errors"
)

const DirectProducerTmpl = `
// {{ .FunctionName }} returns a new {{ .BTypeName }} with the information from
// given {{ .ATypeName }}.
func {{ .FunctionName }}(a {{ .ATypeName }}) {{ .BTypeName }} {
  b := {{ .BTypeNewStatement }}
{{ .Statements }}
  return b
}`

func WithTemplate(t string) FuncOption {
	return func(p *Func) {
		p.Template = t
	}
}

type FuncOption func(p *Func)

func NewFunc(im *packages.Map, tr traverser.GenericTraverser, opts ...FuncOption) *Func {
	f := &Func{
		Imports:   im,
		Traverser: tr,
		Template:  DirectProducerTmpl,
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

type Func struct {
	Imports   *packages.Map
	Traverser traverser.GenericTraverser
	Template  string
}

// Wrap assumes that packages are imported with the same name.
func (f *Func) Wrap(name string, a, b types.Type, extraInput map[string]interface{}) (string, error) {
	content, err := f.Traverser.Print(a, b, "a", "b", 0)
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
	aTypeDec := f.Imports.UseType(an.String())
	aTypeName := fmt.Sprintf("%s%s", aNamePrefix, aTypeDec)
	aNewStatement := fmt.Sprintf("%s{}", aTypeName)
	if aNamePrefix == "*" {
		aNewStatement = fmt.Sprintf("&%s", aNewStatement)
	}
	bTypeDec := f.Imports.UseType(bn.String())
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
	t, err := template.New("func").Parse(f.Template)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, ts)
	return strings.ReplaceAll(result.String(), "\n\n", "\n"), errors.Wrap(err, "cannot execute template")
}
