package wrapper

import (
	"bytes"
	"fmt"
	"github.com/muvaf/typewriter/pkg/imports"
	"io/ioutil"
	"text/template"

	"github.com/pkg/errors"
)

const DefaultFileTmpl = `
{{ .Header }}

package {{ .PackageName }}

import (
{{ .Imports }}
)

{{ .Content }}
`

type DefaultFileTmplInput struct {
	Header      string
	PackageName string
	Imports     string
	Content     string
}

func WithHeaderPath(h string) FileOption {
	return func(f *File) {
		f.HeaderPath = h
	}
}

func WithImports(i imports.Map) FileOption {
	return func(f *File) {
		f.Imports = i
	}
}

type FileOption func(*File)

func NewFile(opts ...FileOption) *File {
	f := &File{
		Imports: imports.Map{},
	}
	for _, fn := range opts {
		fn(f)
	}
	return f
}

type File struct {
	HeaderPath  string
	Imports imports.Map
}

// Wrap writes the objects to the file one by one.
func (f *File) Wrap(packageName string, objects ...string) ([]byte, error) {
	importStatements := ""
	for p, a := range f.Imports {
		// We always use an alias because package name does not necessarily equal
		// to that the last word in the path, hence it's not completely safe to
		// not use an alias even though there is no conflict.
		importStatements += fmt.Sprintf("%s \"%s\"\n", a, p)
	}
	content := ""
	for _, o := range objects {
		content += fmt.Sprintf("%s\n", o)
	}
	header, err := ioutil.ReadFile(f.HeaderPath)
	if err != nil {
		return nil, err
	}
	ts := DefaultFileTmplInput{
		Header:      string(header),
		PackageName: packageName,
		Imports:     importStatements,
		Content:     content,
	}
	t, err := template.New("file").Parse(DefaultFileTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse template")
	}
	result := &bytes.Buffer{}
	err = t.Execute(result, ts)
	return result.Bytes(), errors.Wrap(err, "cannot execute template")
}