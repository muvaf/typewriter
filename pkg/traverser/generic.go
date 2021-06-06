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

	"github.com/muvaf/typewriter/pkg/packages"
)

func WithMap(m MapTraverser) Option {
	return func(g *Generic) {
		m.SetGenericTraverser(g)
		g.Map = m
	}
}

func WithMapTemplate(t string) Option {
	return func(g *Generic) {
		g.Map.SetTemplate(t)
	}
}

func WithPointer(p PointerTraverser) Option {
	return func(g *Generic) {
		p.SetGenericTraverser(g)
		g.Pointer = p
	}
}

func WithPointerTemplate(t string) Option {
	return func(g *Generic) {
		g.Pointer.SetTemplate(t)
	}
}

func WithBasic(b BasicTraverser) Option {
	return func(g *Generic) {
		g.Basic = b
	}
}

func WithBasicTemplate(t map[types.BasicKind]string) Option {
	return func(g *Generic) {
		g.Basic.SetTemplate(t)
	}
}

func WithBasicPointerTemplate(t map[types.BasicKind]string) Option {
	return func(g *Generic) {
		g.Basic.SetPointerTemplate(t)
	}
}

func WithNamed(n NamedTraverser) Option {
	return func(g *Generic) {
		n.SetGenericTraverser(g)
		g.Named = n
	}
}

func WithSlice(s SliceTraverser) Option {
	return func(g *Generic) {
		s.SetGenericTraverser(g)
		g.Slice = s
	}
}

func WithSliceTemplate(t string) Option {
	return func(g *Generic) {
		g.Slice.SetTemplate(t)
	}
}

type Option func(*Generic)

func NewGeneric(im *packages.Imports, opts ...Option) *Generic {
	g := &Generic{
		Imports: im,
		Slice:   NewSlice(im),
		Named:   NewNamed(),
		Basic:   NewBasic(),
		Map:     NewMap(im),
		Pointer: NewPointer(im),
	}
	for _, f := range opts {
		f(g)
	}
	g.Slice.SetGenericTraverser(g)
	g.Map.SetGenericTraverser(g)
	g.Named.SetGenericTraverser(g)
	g.Pointer.SetGenericTraverser(g)
	return g
}

type Generic struct {
	Imports *packages.Imports
	Named   NamedTraverser
	Slice   SliceTraverser
	Basic   BasicTraverser
	Map     MapTraverser
	Pointer PointerTraverser
}

func (g *Generic) Print(a, b types.Type, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	switch at := a.(type) {
	case *types.Pointer:
		bt, ok := b.(*types.Pointer)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		// No need to initialize a new pointer for basic types since the operations
		// are done directly on them, not in their fields as opposed to structs.
		atb, aBasic := at.Elem().(*types.Basic)
		btb, bBasic := bt.Elem().(*types.Basic)
		if aBasic && bBasic {
			o, err := g.Basic.Print(atb, btb, aFieldPath, bFieldPath, true)
			return o, errors.Wrap(err, "cannot traverse basic pointer type")
		}
		// This is to guard for types like `*[]string` that are implicitly pointer
		// of pointers. After processing that it's a pointer, we cannot proceed
		// to process the element type without de-referencing the pointer.
		// TODO(muvaf): This is probably needed for *map[string]string as well.
		// todo(muvaf): THIS COMPILES BUT DEREFERENCES EMPTY POINTER!!!!!! FIX!
		ats, aSlice := at.Elem().(*types.Slice)
		bts, bSlice := bt.Elem().(*types.Slice)
		if aSlice && bSlice {
			o, err := g.Slice.Print(ats, bts, "(*"+aFieldPath+")", "(*"+bFieldPath+")", levelNum)
			return o, errors.Wrap(err, "cannot traverse basic pointer type")
		}
		o, err := g.Pointer.Print(at, bt, aFieldPath, bFieldPath, levelNum)
		return o, errors.Wrap(err, "cannot traverse pointer type")
	case *types.Slice:
		bt, ok := b.(*types.Slice)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		o, err := g.Slice.Print(at, bt, aFieldPath, bFieldPath, levelNum)
		return o, errors.Wrap(err, "cannot traverse slice type")
	case *types.Map:
		bt, ok := b.(*types.Map)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		o, err := g.Map.Print(at, bt, aFieldPath, bFieldPath, levelNum)
		return o, errors.Wrap(err, "cannot traverse map type")
	case *types.Named:
		bt, ok := b.(*types.Named)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		o, err := g.Named.Print(at, bt, aFieldPath, bFieldPath, levelNum)
		return o, errors.Wrap(err, "cannot traverse named type")
	case *types.Basic:
		bt, ok := b.(*types.Basic)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		o, err := g.Basic.Print(at, bt, aFieldPath, bFieldPath, false)
		return o, errors.Wrap(err, "cannot traverse basic type")
	case *types.Struct: // unnamed struct fields.
		return "", nil
	default:
		return "", fmt.Errorf("unknown type in recursion: %s\n", at.String())
	}
}
