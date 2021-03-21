package cmd

import (
	"bytes"
	"go/types"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	packages2 "github.com/muvaf/typewriter/pkg/packages"

	"github.com/muvaf/typewriter/pkg/wrapper"

	"github.com/pkg/errors"
)

type GeneratorChain []Generator

func (gc GeneratorChain) Generate(t *types.Named, cm *packages2.CommentMarkers) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for i, g := range gc {
		if !g.Matches(cm) {
			continue
		}
		out, err := g.Generate(t, cm)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot run generator at index %d", i)
		}
		for k, v := range out {
			result[k] = v
		}
	}
	return result, nil
}

type File struct {
	SourcePackagePath string
	TargetFilePath    string
	FileTemplatePath  string
	LicenseHeaderPath string
	DisableLinter     bool
	NewGeneratorFns   []NewGeneratorFn
	Cache             *packages2.Cache
}

func (f *File) Run() error {
	sourcePkg, err := f.Cache.GetPackage(f.SourcePackagePath)
	if err != nil {
		return errors.Wrap(err, "cannot get source package")
	}
	recipe, err := packages2.LoadCommentMarkers(sourcePkg)
	if err != nil {
		return errors.Wrap(err, "cannot scan comment markers")
	}
	targetPkgPath := f.TargetFilePath[:strings.LastIndex(f.TargetFilePath, "/")]
	targetPkgName := targetPkgPath[strings.LastIndex(targetPkgPath, "/")+1:]
	file := wrapper.NewFile(targetPkgName, f.FileTemplatePath,
		wrapper.WithHeaderPath(f.LicenseHeaderPath),
	)
	gens := GeneratorChain{}
	for _, fn := range f.NewGeneratorFns {
		gens = append(gens, fn(f.Cache, file.Imports))
	}
	input := map[string]interface{}{}
	for sourceType, commentMarker := range recipe {
		generated, err := gens.Generate(sourceType, commentMarker)
		if err != nil {
			return errors.Wrapf(err, "cannot run generators for type %s", sourceType.Obj().Name())
		}
		for k, v := range generated {
			input[k] = v
		}
	}
	if err := os.MkdirAll(targetPkgPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "cannot create target package directory %s", targetPkgPath)
	}
	final, err := file.Wrap(input)
	if err != nil {
		return err
	}
	if !f.DisableLinter {
		fb := bytes.NewBuffer(final)
		cmd := exec.Command("goimports")
		cmd.Stdin = fb
		outb := &bytes.Buffer{}
		cmd.Stdout = outb
		if err := cmd.Run(); err != nil {
			return errors.Wrap(err, "goimports failed")
		}
		final = outb.Bytes()
	}
	return errors.Wrap(ioutil.WriteFile(f.TargetFilePath, final, os.ModePerm), "cannot write to target file path")
}
