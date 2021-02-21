package typewriter

import (
	"fmt"
	"github.com/pkg/errors"
	"go/types"
	"strings"
)

func NewSlice() *Slice {
	return &Slice{}
}

const DefaultFmtSliceWrapper = `
for %s := range %s {
$a
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
	out := fmt.Sprintf(DefaultFmtSliceWrapper, index, bFieldPath)
	statement, err := s.Recursive.Print(a.Elem(), b.Elem(), fmt.Sprintf("%s[%s]", aFieldPath, index), fmt.Sprintf("%s[%s]", bFieldPath, index), levelNum+1)
	return strings.ReplaceAll(out, "$a", statement), errors.Wrap(err, "cannot recursively traverse element type of slice")
}