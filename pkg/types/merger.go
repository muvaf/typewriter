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

func NewMerger(name *types.TypeName, inputTypes []*types.Named, flattener *Flattener) *Merger {
	r := &Merger{
		typeName:   name,
		inputTypes: inputTypes,
		flattener:  flattener,
	}
	return r
}

type Merger struct {
	inputTypes []*types.Named
	typeName   *types.TypeName
	flattener  *Flattener
}

func (m *Merger) Generate() (*types.Named, *packages.CommentMarkers, error) {
	varMap := map[string]*types.Var{}
	tagMap := map[string]string{}
	cm := packages.NewCommentMarkers()
	for _, c := range m.inputTypes {
		addMergedTypeMarker(cm, c)
		in := c.Underlying().(*types.Struct)
		for i := 0; i < in.NumFields(); i++ {
			varMap[in.Field(i).Name()] = in.Field(i)
			tagMap[in.Field(i).Name()] = in.Tag(i)
		}
	}
	fields := make([]*types.Var, len(varMap))
	tags := make([]string, len(varMap))
	i := 0
	for name, v := range varMap {
		fields[i] = v
		if t, ok := tagMap[name]; ok {
			tags[i] = t
		}
		i++
	}
	n := types.NewNamed(m.typeName, types.NewStruct(fields, tags), nil)
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
