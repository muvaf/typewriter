package main

import (
	"path/filepath"

	"github.com/muvaf/typewriter/pkg/packages"

	"github.com/alecthomas/kong"

	"github.com/muvaf/typewriter/pkg/cmd"
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
	f := cmd.File{
		SourcePackagePath: pkgPath,
		TargetFilePath:    filepath.Join(targetPkgPath, "producers.go"),
		// TODO(muvaf): New Go version allows embedding files in the binary. Consider
		// using that
		FileTemplatePath:  "/Users/monus/go/src/github.com/muvaf/typewriter/internal/templates/producers.go.tmpl",
		LicenseHeaderPath: "/Users/monus/go/src/github.com/muvaf/typewriter/internal/header.txt",
		Cache:             c,
		DisableLinter:     disableLinter,
		NewGeneratorFns: []cmd.NewGeneratorFn{
			cmd.NewProducer,
		},
	}
	return f.Run()
}
