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

type FlattenerOption func(*Flattener)

func NewFlattener(im *packages.Imports, opts ...FlattenerOption) *Flattener {
	f := &Flattener{
		Imports: im,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

type Flattener struct {
	Imports *packages.Imports

	TypeFilter  TypeFilter
	FieldFilter FieldFilter
}

func (f *Flattener) Flatten(t *types.Named) map[types.TypeName]types.Type {
	typeMap := map[types.TypeName]types.Type{}
	f.load(typeMap, t)
	return typeMap
}

func (f *Flattener) load(m map[types.TypeName]types.Type, t *types.Named) {
	t = f.TypeFilter.Filter(t)
	if t == nil {
		return
	}
	s, ok := t.Underlying().(*types.Struct)
	if t.Underlying() != nil && !ok {
		// TODO(muvaf): If the underlying type is not Struct, it means it's
		// likely enum, which doesn't have fields. However, if it points to a named
		// type instead of a basic one, we're skipping it.

		// todo: naming collisions? is it possible this function runs with multiple
		// packages?
		m[*t.Obj()] = t.Underlying()
		return
	}
	var fields []*types.Var
	var tags []string
	for i := 0; i < s.NumFields(); i++ {
		field, tag := f.FieldFilter.Filter(s.Field(i), s.Tag(i))
		if field == nil {
			continue
		}
		fields = append(fields, field)
		tags = append(tags, tag)
		ft := field.Type()
		switch u := ft.(type) {
		case *types.Pointer:
			n, ok := u.Elem().(*types.Named)
			if !ok {
				continue
			}
			f.load(m, n)
		case *types.Slice:
			switch n := u.Elem().(type) {
			case *types.Named:
				f.load(m, n)
			case *types.Pointer:
				pn, ok := n.Elem().(*types.Named)
				if !ok {
					continue
				}
				f.load(m, pn)
			}
		case *types.Named:
			f.load(m, u)
		}
	}
	m[*t.Obj()] = types.NewStruct(fields, tags)
}
