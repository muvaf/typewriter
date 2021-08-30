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

package packages

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

const (
	LoadMode = packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax
)

func NewCache(init ...*packages.Package) *Cache {
	s := map[string]*packages.Package{}
	for _, p := range init {
		s[p.PkgPath] = p
	}
	return &Cache{
		store: s,
	}
}

type Cache struct {
	store map[string]*packages.Package
}

// GetTypeWithFullPath returns the type information of the type in given path. The expected
// format is "<package path>.<type name>".
func (pc *Cache) GetTypeWithFullPath(fullPath string) (*types.Named, error) {
	path, name := fullPath[:strings.LastIndex(fullPath, ".")], fullPath[strings.LastIndex(fullPath, ".")+1:]
	return pc.GetType(path, name)
}

// GetType returns the type information of the type in given path. The expected
// format is "<package path>.<type name>".
func (pc *Cache) GetType(packagePath, name string) (*types.Named, error) {
	p, err := pc.GetPackage(packagePath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get package")
	}
	o := p.Types.Scope().Lookup(name)
	if o == nil {
		return nil, errors.Errorf("cannot find given type %s in package %s", name, packagePath)
	}
	if n, ok := o.Type().(*types.Named); ok {
		return n, nil
	}
	return nil, errors.Errorf("type %s is not a named struct", name)
}

// GetPackage accepts absolute local path or Go module path and returns a single
// package. It caches by Go module path in both cases.
func (pc *Cache) GetPackage(absolutePath string) (*packages.Package, error) {
	// Path could be local path or module path but we cache with module path.
	// Since local path has the module path as suffix albeit we don't really know
	// the seperator, we can iterate and find any package whose path has given path
	// as suffix.
	for p, pkg := range pc.store {
		if strings.HasSuffix(absolutePath, p) {
			return pkg, nil
		}
	}
	pkgs, err := packages.Load(&packages.Config{Mode: LoadMode}, absolutePath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load packages in %s", absolutePath)
	}
	for _, pkg := range pkgs {
		if strings.HasSuffix(absolutePath, pkg.PkgPath) {
			if len(pkg.Errors) != 0 {
				errStr := ""
				for _, e := range pkg.Errors {
					errStr += fmt.Sprintf("%s ", e.Error())
				}
				return nil, errors.Errorf("cannot load package with error: %s", errStr)
			}
			pc.store[pkg.PkgPath] = pkg
			return pkg, nil
		}
	}
	return nil, errors.Errorf("cannot find package in %s", absolutePath)
}
