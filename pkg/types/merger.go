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

// TODO(muvaf): Using the result of union operation as ignore func parameter
// could be helpful. Consider providing functions to make this easy. For example,
// `ignore all fields in this type that already exists in that other type.`

func WithIgnore(f ...IgnoreFieldFn) Option {
	return func(rc *Merger) {
		rc.filter.ignore = f
	}
}

type Option func(*Merger)

func NewMerger(name *types.TypeName, inputTypes []*types.Named, opts ...Option) *Merger {
	r := &Merger{
		TypeName:   name,
		InputTypes: inputTypes,
	}
	for _, f := range opts {
		f(r)
	}
	return r
}

type IgnoreFieldFn func(*types.Var) bool

type IgnoreFieldChain []IgnoreFieldFn

func (i IgnoreFieldChain) ShouldIgnore(v *types.Var) bool {
	for _, f := range i {
		if f(v) {
			return true
		}
	}
	return false
}

type filter struct {
	ignore IgnoreFieldChain
}

type Merger struct {
	filter filter

	InputTypes []*types.Named
	TypeName   *types.TypeName
}

func (m *Merger) Generate() (*types.Named, *packages.CommentMarkers, error) {
	varMap := map[string]*types.Var{}
	cm := packages.NewCommentMarkers()
	for _, c := range m.InputTypes {
		addMergedTypeMarker(cm, c)
		cre := c.Underlying().(*types.Struct)
		for i := 0; i < cre.NumFields(); i++ {
			if m.filter.ignore.ShouldIgnore(cre.Field(i)) {
				continue
			}
			varMap[cre.Field(i).Name()] = cre.Field(i)
		}
	}
	fields := make([]*types.Var, len(varMap))
	i := 0
	for _, v := range varMap {
		fields[i] = v
		i++
	}
	n := types.NewNamed(m.TypeName, types.NewStruct(fields, nil), nil)
	return n, cm, nil
}

func addMergedTypeMarker(cm *packages.CommentMarkers, n *types.Named) {
	fullPath := packages.FullPath(n)
	for _, ag := range cm.SectionContents[packages.SectionMerged] {
		if ag == fullPath {
			return
		}
	}
	cm.SectionContents[packages.SectionMerged] = append(cm.SectionContents[packages.SectionMerged], fullPath)
}
