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

package types

import (
	"fmt"
	"go/token"
	"go/types"

	"github.com/muvaf/typewriter/pkg/packages"
)

type TypeFilter interface {
	Filter(*types.Named) *types.Named
}

type FieldFilter interface {
	Filter(field *types.Var, tag string) (*types.Var, string)
}

func WithTypeFilters(tf ...TypeFilter) FlattenerOption {
	return func(f *Flattener) {
		f.TypeFilter = TypeFilterChain(tf)
	}
}

func WithFieldFilters(ff ...FieldFilter) FlattenerOption {
	return func(f *Flattener) {
		f.FieldFilter = FieldFilterChain(ff)
	}
}

func WithRemotePkgPath(path string) FlattenerOption {
	return func(f *Flattener) {
		f.RemotePkgPath = path
	}
}

func WithLocalPkg(pkg *types.Package) FlattenerOption {
	return func(f *Flattener) {
		f.LocalPkg = pkg
	}
}

type FlattenerOption func(*Flattener)

func NewFlattener(im *packages.Imports, opts ...FlattenerOption) *Flattener {
	f := &Flattener{
		Imports:     im,
		TypeFilter:  NopTypeFilter{},
		FieldFilter: NopFieldFilter{},
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

type Flattener struct {
	Imports *packages.Imports
	// RemotePkgPath is the path of the package of the remote type we're pulling
	// the types from.
	RemotePkgPath string
	LocalPkg      *types.Package

	TypeFilter  TypeFilter
	FieldFilter FieldFilter
}

func (f *Flattener) Flatten(t *types.Named) []*types.Named {
	typeMap := map[types.TypeName]*types.Named{}
	f.load(typeMap, t)
	result := make([]*types.Named, len(typeMap))
	i := 0
	for _, n := range typeMap {
		result[i] = n
		i++
	}
	return result
}

func (f *Flattener) load(m map[types.TypeName]*types.Named, t *types.Named) {
	t = f.TypeFilter.Filter(t)
	if t == nil {
		return
	}
	s, ok := t.Underlying().(*types.Struct)
	if !ok {
		// TODO(muvaf): If the underlying type is not Struct, it means it's
		// likely enum, which doesn't have fields, hence no field iteration.
		// However, if it points to a named type instead of a basic one, we're
		// skipping it.
		// TODO(muvaf): naming collisions? is it possible this function runs
		// with multiple packages?
		b, ok := t.Underlying().(*types.Basic)
		if !ok {
			fmt.Printf("only types whose underlying is struct or basic are supported, skipping %s\n", t.Obj().Name())
		}
		ntn := types.NewTypeName(token.NoPos, f.LocalPkg, t.Obj().Name(), nil)
		methods := make([]*types.Func, t.NumMethods())
		for j := 0; j < t.NumMethods(); j++ {
			methods[j] = t.Method(j)
		}
		m[*ntn] = types.NewNamed(ntn, b, methods)
		return
	}
	var fields []*types.Var
	var tags []string
	for i := 0; i < s.NumFields(); i++ {
		// TODO(muvaf): Make this optional.
		if !s.Field(i).Exported() {
			continue
		}
		field, tag := f.FieldFilter.Filter(s.Field(i), s.Tag(i))
		if field == nil {
			continue
		}
		switch u := field.Type().(type) {
		case *types.Pointer:
			newElem := u.Elem()
			if n, ok := u.Elem().(*types.Named); ok {
				f.load(m, n)
				if n.Obj().Pkg().Path() == f.RemotePkgPath {
					newElem = NewNamedInLocalPkg(n, f.LocalPkg)
				}
			}
			field = types.NewField(field.Pos(), f.LocalPkg, field.Name(), types.NewPointer(newElem), field.Embedded())
		case *types.Slice:
			newElem := u.Elem()
			switch n := u.Elem().(type) {
			case *types.Named:
				f.load(m, n)
				if n.Obj().Pkg().Path() == f.RemotePkgPath {
					newElem = NewNamedInLocalPkg(n, f.LocalPkg)
				}
			case *types.Pointer:
				if pn, ok := n.Elem().(*types.Named); ok {
					f.load(m, pn)
					if pn.Obj().Pkg().Path() == f.RemotePkgPath {
						newElem = types.NewPointer(NewNamedInLocalPkg(pn, f.LocalPkg))
					}
				}
			}
			field = types.NewField(field.Pos(), f.LocalPkg, field.Name(), types.NewSlice(newElem), field.Embedded())
		case *types.Named:
			newNamed := u
			f.load(m, u)
			if u.Obj().Pkg().Path() == f.RemotePkgPath {
				newNamed = NewNamedInLocalPkg(u, f.LocalPkg)
			}
			field = types.NewField(field.Pos(), f.LocalPkg, field.Name(), newNamed, field.Embedded())
		default:
			field = types.NewField(field.Pos(), f.LocalPkg, field.Name(), field.Type(), field.Embedded())
		}
		fields = append(fields, field)
		tags = append(tags, tag)
	}
	ns := types.NewStruct(fields, tags)
	ntn := types.NewTypeName(token.NoPos, f.LocalPkg, t.Obj().Name(), nil)
	methods := make([]*types.Func, t.NumMethods())
	for j := 0; j < t.NumMethods(); j++ {
		methods[j] = t.Method(j)
	}
	m[*ntn] = types.NewNamed(ntn, ns, methods)
}

func NewNamedInLocalPkg(t *types.Named, pkg *types.Package) *types.Named {
	ntn := types.NewTypeName(t.Obj().Pos(), pkg, t.Obj().Name(), nil)
	methods := make([]*types.Func, t.NumMethods())
	for j := 0; j < t.NumMethods(); j++ {
		methods[j] = t.Method(j)
	}
	return types.NewNamed(ntn, t.Underlying(), methods)
}
