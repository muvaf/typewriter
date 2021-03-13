package wrapper

import (
	"bytes"
	"fmt"
	"github.com/muvaf/typewriter/pkg/imports"
	"go/types"
	"text/template"

	"github.com/pkg/errors"
)

const DefaultFuncTmpl = `
// {{ .FunctionName }} returns a new {{ .BTypeName }} with the information from
// given {{ .ATypeName }}.
func {{ .FunctionName }}(a {{ .ATypeName }}) {{ .BTypeName }} {
  b := {{ .BTypeNewStatement }}
{{ .Content }}
  return b
}`

type DefaultFuncTmplInput struct {
	FunctionName      string
	ATypeName         string
	ATypeNewStatement string
	BTypeName         string
	BTypeNewStatement string
	Content           string
}

func NewFunc(im imports.Map) *Func {
	return &Func{
		Imports: im,
	}
}

type Func struct{
	Imports imports.Map
}

// Wrap assumes that packages are imported with the same name.
func (f *Func) Wrap(a, b types.Type, name, content string) (string, error) {
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
	ts := DefaultFuncTmplInput{
		FunctionName:      name,
		ATypeName:         aTypeName,
		ATypeNewStatement: aNewStatement,
		BTypeName:         bTypeName,
		BTypeNewStatement: bNewStatement,
		Content:           content,
	}
	t, err := template.New("func").Parse(DefaultFuncTmpl)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, ts)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
