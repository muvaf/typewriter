package traverser

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

const AssignmentTmpl = `
{{ .BFieldPath }} = {{ .AFieldPath }}`

type AssignmentTmplInput struct {
	AFieldPath string
	BFieldPath string
}

func NewBasic() *Basic {
	b := &Basic{
		Templates: map[types.BasicKind]string{},
	}
	for i := 1; i < 26; i++ {
		b.Templates[types.BasicKind(i)] = AssignmentTmpl
	}
	return b
}

type Basic struct {
	Templates map[types.BasicKind]string
}

func (bs *Basic) SetTemplate(t map[types.BasicKind]string) {
	bs.Templates = t
}

func (bs *Basic) Print(a, b *types.Basic, aFieldPath, bFieldPath string) (string, error) {
	if a.Kind() != b.Kind() {
		return "", fmt.Errorf(errFmtNotSameKind, a.String(), b.String())
	}
	tmpl, ok := bs.Templates[a.Kind()]
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
