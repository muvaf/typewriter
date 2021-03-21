package imports

import (
	"fmt"
	"strings"
)

func NewMap(selfPackage string) *Map {
	return &Map{
		Package: selfPackage,
		Imports: map[string]string{},
	}
}

type Map struct {
	Package string
	Imports map[string]string
}

// TODO(muvaf): We could make this routine-safe but it's not necessary for now.

// UseType adds the package to the import map and returns the alias you
// can use in that Go file.
func (m *Map) UseType(in string) string {
	pkg, typeNameFmt := parseTypeDec(in)
	if isBuiltIn(typeNameFmt) {
		return in
	}
	if pkg == m.Package {
		// this is a temp hack for my own code :(
		return strings.ReplaceAll(typeNameFmt, "%s.", "")
	}
	val, ok := m.Imports[pkg]
	if ok {
		return fmt.Sprintf(typeNameFmt, val)
	}
	tmp := map[string]struct{}{}
	for _, a := range m.Imports {
		tmp[a] = struct{}{}
	}
	words := strings.Split(pkg, "/")
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
	m.Imports[pkg] = alias
	return fmt.Sprintf(typeNameFmt, alias)
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
