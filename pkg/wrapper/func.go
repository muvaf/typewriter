package wrapper

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

const DefaultGenerator = `
func {{ .FunctionName }}(a {{ .ATypeName }}) {{ .BTypeName }} {
  b := {{ .BTypeNewStatement }}
  {{ .Content }}
  return b
}
`

type DefaultGeneratorStruct struct {
	FunctionName      string
	ATypeName         string
	BTypeName         string
	BTypeNewStatement string
	Content           string
}

func NewFunc() *Func {
	return &Func{}
}

type Func struct{}

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
	anp := strings.Split(an.Obj().Pkg().Name(), "/")
	aTypeName := fmt.Sprintf("%s%s.%s", aNamePrefix, anp[len(anp)-1], an.Obj().Name())
	aNewStatement := fmt.Sprintf("%s{}", aTypeName)
	if aNamePrefix == "*" {
		aNewStatement = fmt.Sprintf("&%s", aNewStatement)
	}
	bnp := strings.Split(bn.Obj().Pkg().Name(), "/")
	bTypeName := fmt.Sprintf("%s%s.%s", bNamePrefix, bnp[len(anp)-1], an.Obj().Name())
	bNewStatement := fmt.Sprintf("%s{}", bTypeName)
	if bNamePrefix == "*" {
		bNewStatement = fmt.Sprintf("&%s", bNewStatement)
	}
	ts := DefaultGeneratorStruct{
		FunctionName:      name,
		ATypeName:         aTypeName,
		BTypeName:         bTypeName,
		BTypeNewStatement: bNewStatement,
		Content:           content,
	}
	t, err := template.New("func").Parse(DefaultGenerator)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, ts)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
