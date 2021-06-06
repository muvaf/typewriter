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
		Templates:        map[types.BasicKind]string{},
		PointerTemplates: map[types.BasicKind]string{},
	}
	for i := 1; i < 26; i++ {
		b.Templates[types.BasicKind(i)] = AssignmentTmpl
		b.PointerTemplates[types.BasicKind(i)] = AssignmentTmpl
	}
	return b
}

type Basic struct {
	Templates        map[types.BasicKind]string
	PointerTemplates map[types.BasicKind]string
}

func (bs *Basic) SetTemplate(t map[types.BasicKind]string) {
	bs.Templates = t
}

func (bs *Basic) SetPointerTemplate(t map[types.BasicKind]string) {
	bs.PointerTemplates = t
}

func (bs *Basic) Print(a, b *types.Basic, aFieldPath, bFieldPath string, isPointer bool) (string, error) {
	if a.Kind() != b.Kind() {
		return "", fmt.Errorf(errFmtNotSameKind, a.String(), b.String())
	}
	tmplStore := bs.Templates
	if isPointer {
		tmplStore = bs.PointerTemplates
	}
	tmpl, ok := tmplStore[a.Kind()]
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
