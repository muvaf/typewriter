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

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/cmd"
	"github.com/muvaf/typewriter/pkg/packages"
	"github.com/muvaf/typewriter/pkg/wrapper"
)

type typeWriterCLI struct {
	PackagePath       string `help:"Path to package dir to scan" type:"path" required:""`
	TargetPackagePath string `help:"Package to write the generated files. If not given, package path will be used." type:"path"`
	DisableLinter     bool   `help:"Option to disable linting the output. Useful for debugging errors."`
}

func main() {
	cli := &typeWriterCLI{}
	ctx := kong.Parse(cli)
	targetPackagePath := cli.TargetPackagePath
	if targetPackagePath == "" {
		targetPackagePath = cli.PackagePath
	}
	ctx.FatalIfErrorf(PrintProducers(cli.PackagePath, cli.TargetPackagePath, cli.DisableLinter), "cannot print producers")
}

func PrintProducers(pkgPath, targetPkgPath string, disableLinter bool) error {
	c := packages.NewCache()
	targetPkgName := targetPkgPath[strings.LastIndex(targetPkgPath, "/")+1:]
	tmplPath := "/Users/monus/go/src/github.com/muvaf/typewriter/internal/templates/producers.go.tmpl"
	headerPath := "/Users/monus/go/src/github.com/muvaf/typewriter/internal/header.txt"
	file := wrapper.NewFile(targetPkgName, tmplPath,
		wrapper.WithHeaderPath(headerPath),
	)
	vars := map[string]interface{}{}
	f := cmd.NewFunctions(c, file.Imports, pkgPath,
		cmd.WithNewFuncGeneratorFns(cmd.NewProducers))
	fns, err := f.Run()
	if err != nil {
		return err
	}
	for k, v := range fns {
		vars[k] = v
	}
	final, err := file.Wrap(vars)
	if err != nil {
		return err
	}
	if !disableLinter {
		fb := bytes.NewBuffer(final)
		command := exec.Command("goimports")
		command.Stdin = fb
		outb := &bytes.Buffer{}
		command.Stdout = outb
		if err := command.Run(); err != nil {
			return errors.Wrap(err, "goimports failed")
		}
		final = outb.Bytes()
	}
	if err := os.MkdirAll(targetPkgPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "cannot create target package directory %s", targetPkgPath)
	}
	return errors.Wrap(ioutil.WriteFile(filepath.Join(targetPkgPath, "producers.go"), final, os.ModePerm), "cannot write to target file path")
}
