package imports

import "strings"

type Map map[string]string

// TODO(muvaf): We could make this routine-safe but it's not necessary for now.

// UsePackage adds the package to the import map and returns the alias you
// can use in that Go file.
func (m Map) UsePackage(path string) string {
	val, ok := m[path]
	if ok {
		return val
	}
	tmp := map[string]struct{}{
		"": {},
	}
	for _, a := range m {
		tmp[a] = struct{}{}
	}
	words := strings.Split(path, "/")
	result := ""
	for _, w := range words {
		if _, ok := tmp[result]; !ok {
			break
		}
		result += strings.ReplaceAll(w, ".", "")
	}
	// Because the main map guarantees to have each of its entry to be different,
	// the for loop above has to find a meaningful result before running out.
	// The ReplaceAll statement is pinching hole in this completeness, but considering
	// the paths are URLs, replacing dot with nothing should be fine.
	m[path] = result
	return result
}
