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
	"strings"
)

func NewImports(pkgPath, pkgName string) *Imports {
	return &Imports{
		PackagePath: pkgPath,
		PackageName: pkgName,
		Imports:     map[string]string{},
	}
}

type Imports struct {
	PackagePath string
	PackageName string
	Imports     map[string]string
}

// TODO(muvaf): We could make this routine-safe but it's not necessary for now.

// UseType adds the package of given type to the import map and returns the alias
// you can use in that Go file.
func (m *Imports) UseType(in string) string {
	pkgPath, typeNameFmt := parseTypeDec(in)
	if isBuiltIn(typeNameFmt) {
		return in
	}
	if strings.HasSuffix(m.PackagePath, pkgPath) {
		// this is a temp hack for my own code :(
		return strings.ReplaceAll(typeNameFmt, "%s.", "")
	}
	val, ok := m.Imports[pkgPath]
	if ok {
		return fmt.Sprintf(typeNameFmt, val)
	}
	tmp := map[string]struct{}{}
	for _, a := range m.Imports {
		tmp[a] = struct{}{}
	}
	words := strings.Split(pkgPath, "/")
	alias := words[len(words)-1]
	for i := len(words) - 2; i >= 0; i-- {
		if _, ok := tmp[alias]; !ok {
			break
		}
		alias += strings.ReplaceAll(words[i], ".", "")
	}
	// Because the main map guarantees to have each of its entry to be different,
	// the for loop above has to find a meaningful result before running out.
	// The ReplaceAll statement is pinching hole in this completeness, but considering
	// the paths are URLs, replacing dot with nothing should be fine.
	m.Imports[pkgPath] = alias
	return fmt.Sprintf(typeNameFmt, alias)
}

// UsePackage adds the package to the import map and returns the alias you
// can use in that Go file. The returned package name will have "." as suffix
// if it is from a different package than the current one.
func (m *Imports) UsePackage(pkgPath string) string {
	if strings.HasSuffix(m.PackagePath, pkgPath) {
		return ""
	}
	val, ok := m.Imports[pkgPath]
	if ok {
		return val + "."
	}
	tmp := map[string]struct{}{}
	for _, a := range m.Imports {
		tmp[a] = struct{}{}
	}
	words := strings.Split(pkgPath, "/")
	alias := words[len(words)-1]
	for i := len(words) - 2; i >= 0; i-- {
		if _, ok := tmp[alias]; !ok {
			break
		}
		alias += strings.ReplaceAll(words[i], ".", "")
	}
	// Because the main map guarantees to have each of its entry to be different,
	// the for loop above has to find a meaningful result before running out.
	// The ReplaceAll statement is pinching hole in this completeness, but considering
	// the paths are URLs, replacing dot with nothing should be fine.
	m.Imports[pkgPath] = alias
	return alias + "."
}

// parseTypeDec returns the full package name and the type that can be used in
// the code. You need to use formatter to replace %s in the type name with alias
// that's used.
func parseTypeDec(s string) (string, string) {
	// It is compatible with:
	// []pkg.Type
	// []*pkg.Type
	// pkg.Type
	// *pkg.Type
	// Get rid of slice and pointer chars.
	tmp := strings.NewReplacer(
		"[", "",
		"]", "",
		"*", "").Replace(s)
	dotIndex := strings.LastIndex(tmp, ".")
	if dotIndex == -1 {
		return "", s
	}
	pkgName := tmp[:dotIndex]
	return pkgName, strings.ReplaceAll(s, pkgName, "%s")
}

// TODO(muvaf): find a better method to check this.
func isBuiltIn(s string) bool {
	s = strings.NewReplacer("*", "", "[]", "").Replace(s)
	switch s {
	case "bool", "string", "int", "int64", "map[string]string":
		return true
	}
	return false
}
