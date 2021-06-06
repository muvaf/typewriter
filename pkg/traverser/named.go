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
	Generic GenericTraverser
}

func (s *Named) SetGenericTraverser(p GenericTraverser) {
	s.Generic = p
}

func (s *Named) Print(a, b *types.Named, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	// TODO(muvaf): This could be *types.Map and valid.
	at, ok := a.Underlying().(*types.Struct)
	if !ok {
		return "", nil
	}
	bt := b.Underlying().(*types.Struct)
	if !ok {
		return "", nil
	}
	out := ""
	for i := 0; i < at.NumFields(); i++ {
		if at.Field(i).Name() == "_" {
			continue
		}
		// TODO(muvaf): make this default but modifiable in the future.
		if !at.Field(i).Exported() {
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
		add, err := s.Generic.Print(af.Type(), bf.Type(), fmt.Sprintf("%s.%s", aFieldPath, af.Name()), fmt.Sprintf("%s.%s", bFieldPath, bf.Name()), levelNum)
		if err != nil {
			return "", errors.Wrap(err, "cannot recursively traverse field of named type")
		}
		out += add
	}
	return out, nil
}
