package traverser

import (
	"fmt"
	"go/types"

	"github.com/pkg/errors"
)

func NewNamed() *Named {
	return &Named{}
}

type Named struct {
	Recursive TypeTraverser
}

func (s *Named) SetTypeTraverser(p TypeTraverser) {
	s.Recursive = p
}

func (s *Named) Print(a, b *types.Named, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	at := a.Underlying().(*types.Struct)
	bt := b.Underlying().(*types.Struct)
	out := ""
	for i := 0; i < at.NumFields(); i++ {
		if at.Field(i).Name() == "_" {
			continue
		}
		af := at.Field(i)
		var bf *types.Var
		for j := 0; j < bt.NumFields(); j++ {
			if bt.Field(j).Name() == af.Name() {
				bf = bt.Field(j)
				break
			}
		}
		if bf == nil {
			continue
		}
		add, err := s.Recursive.Print(af.Type(), bf.Type(), fmt.Sprintf("%s.%s", aFieldPath, af.Name()), fmt.Sprintf("%s.%s", bFieldPath, bf.Name()), levelNum)
		if err != nil {
			return "", errors.Wrap(err, "cannot recursively traverse field of named type")
		}
		out += add
	}
	return out, nil
}
