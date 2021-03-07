package typewriter

import (
	"bytes"
	"fmt"
	"go/types"
	"text/template"

	"github.com/pkg/errors"
)

const (
	errFmtNotSameKind = "not same basic kind: %s and %s"
	errFmtUnknownKind = "unknown basic kind: %s"
)

func WithTmpl(kind types.BasicKind, tmpl string) BasicOption {
	return func(b *Basic) {
		b.Templates[kind] = tmpl
	}
}

type BasicOption func(*Basic)

const (
	AssignmentTmpl = `
{{ .AFieldPath }} = {{ .BFieldPath }}`
)

type AssignmentTmplInput struct {
	AFieldPath string
	BFieldPath string
}

func NewBasic(opts ...BasicOption) *Basic {
	b := &Basic{
		Templates: map[types.BasicKind]string{},
	}
	for i := 1; i < 26; i++ {
		b.Templates[types.BasicKind(i)] = AssignmentTmpl
	}
	for _, f := range opts {
		f(b)
	}
	return b
}

type Basic struct {
	Templates map[types.BasicKind]string
}

func (s *Basic) Print(a, b *types.Basic, aFieldPath, bFieldPath string) (string, error) {
	if a.Kind() != b.Kind() {
		return "", fmt.Errorf(errFmtNotSameKind, a.String(), b.String())
	}
	tmpl, ok := s.Templates[a.Kind()]
	if !ok {
		return "", fmt.Errorf(errFmtUnknownKind, a.String())
	}
	i := AssignmentTmplInput{
		AFieldPath: aFieldPath,
		BFieldPath: bFieldPath,
	}
	t, err := template.New("basic").Parse(tmpl)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, i)
	return string(result.Bytes()), errors.Wrap(err, "cannot execute template")
}
