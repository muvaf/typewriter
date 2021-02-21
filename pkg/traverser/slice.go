package typewriter

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/pkg/errors"
)

func NewSlice() *Slice {
	return &Slice{}
}

const DefaultFmtSliceWrapper = `
if len($a) != 0 {
  $b := make($typeb, len($a))
  for $index := range $a {
    $statement
  }
}
`

type Slice struct {
	Recursive TypeTraverser
}

func (s *Slice) SetTypeTraverser(p TypeTraverser) {
	s.Recursive = p
}

func (s *Slice) Print(a, b *types.Slice, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	index := fmt.Sprintf("v%d", levelNum)
	statement, err := s.Recursive.Print(a.Elem(), b.Elem(), fmt.Sprintf("%s[%s]", aFieldPath, index), fmt.Sprintf("%s[%s]", bFieldPath, index), levelNum+1)
	if err != nil {
		return "", errors.Wrap(err, "cannot recursively traverse element type of slice")
	}
	out := strings.ReplaceAll(DefaultFmtSliceWrapper, "$index", index)
	out = strings.ReplaceAll(out, "$statement", statement)
	out = strings.ReplaceAll(out, "$a", aFieldPath)
	out = strings.ReplaceAll(out, "$b", bFieldPath)
	out = strings.ReplaceAll(out, "$typeb", b.String())
	return out, nil
}
