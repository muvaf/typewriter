package main

import (
	"bytes"
	"fmt"
	"go/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"

	"github.com/muvaf/typewriter/pkg/imports"
	"github.com/muvaf/typewriter/pkg/scanner"
	"github.com/muvaf/typewriter/pkg/traverser"
	"github.com/muvaf/typewriter/pkg/wrapper"
)

const (
	LoadMode = packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax
)

type typeWriterCLI struct {
	PackagePath       string `help:"Path to package dir to scan" type:"path" required:""`
	TargetPackagePath string `help:"Package to write the generated files. If not given, package path will be used." type:"path"`
}

func main() {
	cli := &typeWriterCLI{}
	ctx := kong.Parse(cli)
	ctx.FatalIfErrorf(PrintConversions(cli.PackagePath, cli.TargetPackagePath), "cannot print conversions")
}

func PrintConversions(pkgPath, targetPkgPath string) error {
	p, err := loadPackage(pkgPath)
	if err != nil {
		return errors.Wrap(err, "cannot load package")
	}
	pkgCache := map[string]*packages.Package{
		pkgPath: p,
	}
	menu := scanner.ScanCommentTags(p)
	var result []string
	// TODO(muvaf): Target package may have a different name other than folder name.
	targetPkgName := getPkgNameFromPath(targetPkgPath)
	fl := wrapper.NewFile(targetPkgName, "/Users/monus/go/src/github.com/muvaf/typewriter/internal/templates/conversions.go.tmpl",
		wrapper.WithHeaderPath("/Users/monus/go/src/github.com/crossplane/provider-gcp/hack/boilerplate.go.txt"),
	)
	for source, ct := range menu {
		for agType := range ct.GetAggregatedTypes() {
			remotePkgPath, remoteTypeName := parseTypePath(agType)
			p, ok := pkgCache[remotePkgPath]
			if !ok {
				pkgCache[remotePkgPath], err = loadPackage(remotePkgPath)
				if err != nil {
					return err
				}
				p = pkgCache[remotePkgPath]
			}
			remoteType := p.Types.Scope().Lookup(remoteTypeName)
			if remoteType == nil {
				return errors.Errorf("cannot find type %s in package %s", remoteType, remotePkgPath)
			}
			remoteNamed, ok := remoteType.Type().(*types.Named)
			if !ok {
				return errors.Errorf("remote type is not a named struct")
			}
			generated, err := getConversion(fl.Imports, source, remoteNamed)
			if err != nil {
				return errors.Wrapf(err, "cannot write conversion from %s to %s", source.Obj().Name(), remoteNamed.Obj().Name())
			}
			result = append(result, generated)
		}
	}
	if len(result) == 0 {
		return nil
	}
	out := ""
	for _, function := range result {
		out += fmt.Sprintf("%s\n", function)
	}
	if err := os.MkdirAll(targetPkgPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "cannot create target package directory %s", targetPkgPath)
	}
	input := map[string]interface{}{
		"Functions": out,
	}
	file, err := fl.Wrap(input)
	if err != nil {
		return err
	}
	fb := bytes.NewBuffer(file)
	cmd := exec.Command("goimports")
	cmd.Stdin = fb
	outb := &bytes.Buffer{}
	cmd.Stdout = outb
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "goimports failed")
	}
	if err := ioutil.WriteFile(filepath.Join(targetPkgPath, "conversions.go"), outb.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func loadPackage(path string) (*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{Mode: LoadMode}, path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load packages in %s", path)
	}
	for _, pkg := range pkgs {
		if strings.HasSuffix(path, pkg.PkgPath) {
			if len(pkg.Errors) != 0 {
				errStr := ""
				for _, e := range pkg.Errors {
					errStr += fmt.Sprintf("%s ", e.Error())
				}
				return nil, errors.Errorf("cannot load package with error: %s", errStr)
			}
			return pkg, nil
		}
	}
	return nil, errors.Errorf("cannot find package in %s", path)
}

func getConversion(im *imports.Map, source, target *types.Named) (string, error) {
	fn := wrapper.NewFunc(im, traverser.NewGeneric(im))
	generated, err := fn.Wrap(fmt.Sprintf("Generate%s", target.Obj().Name()), source, target, nil)
	if err != nil {
		return "", errors.Wrap(err, "cannot wrap function")
	}
	return generated, nil
}

func getPkgNameFromPath(s string) string {
	l := strings.Split(s, "/")
	return l[len(l)-1]
}

func parseTypePath(t string) (path string, name string) {
	return t[:strings.LastIndex(t, ".")], t[strings.LastIndex(t, ".")+1:]
}
