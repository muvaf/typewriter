package typewriter

import (
	"fmt"
	"go/types"
)

func NewNamed() *Named {
	return &Named{}
}

type Named struct {
	Recursive GeneralPrinter
}

func (s *Named) SetGeneralPrinter(p GeneralPrinter) {
	s.Recursive = p
}

func (s *Named) Print(a, b *types.Named, aFieldPath, bFieldPath string, levelNum int) string {
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
		fmt.Println("af type " + af.Type().String())
		fmt.Println("af name " + af.Name())
		fmt.Println("bf type " + bf.Type().String())
		fmt.Println("bf name " + bf.Name())
		out += s.Recursive.Print(af.Type(), bf.Type(), fmt.Sprintf("%s.%s", aFieldPath, af.Name()), fmt.Sprintf("%s.%s", bFieldPath, bf.Name()), levelNum)
	}
	return out
}
