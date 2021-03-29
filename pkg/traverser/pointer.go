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
	Type     GenericTraverser
}

func (p *Pointer) SetTemplate(t string) {
	p.Template = t
}

func (p *Pointer) SetGenericTraverser(tt GenericTraverser) {
	p.Type = tt
}

func (p *Pointer) Print(a, b *types.Pointer, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	statements, err := p.Type.Print(a.Elem(), b.Elem(), aFieldPath, bFieldPath, levelNum)
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
