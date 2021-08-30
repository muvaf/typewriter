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

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

func NewTraverser(cache *packages.Cache, typeProcessors TypeProcessorChain, fieldProcessors FieldProcessorChain) *Traverser {
	return &Traverser{
		cache:           cache,
		commentCache:    packages.NewCommentCache(cache),
		TypeProcessors:  typeProcessors,
		FieldProcessors: fieldProcessors,
	}
}

type Traverser struct {
	TypeProcessors  TypeProcessorChain
	FieldProcessors FieldProcessorChain

	cache        *packages.Cache
	commentCache *packages.CommentCache
}

func (t *Traverser) Traverse(n *types.Named, fieldPath ...string) error {
	pComments, err := t.commentCache.GetComments(n.Obj().Pkg().Path())
	if err != nil {
		return errors.Wrapf(err, "cannot get comments for package %s", n.Obj().Pkg().Path())
	}
	if err := t.TypeProcessors.Process(n, pComments[n.Obj()]); err != nil {
		return errors.Wrapf(err, "type processors failed to run for type %s", n.Obj().Name())
	}
	st, ok := n.Underlying().(*types.Struct)
	if !ok {
		return nil
	}
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		tag := st.Tag(i)
		if err := t.FieldProcessors.Process(n, field, tag, pComments[field], fieldPath); err != nil {
			return errors.Wrapf(err, "field processors failed to run for field %s of type %s", field.Name(), n.Obj().Name())
		}
		ft, ok := field.Type().(*types.Named)
		if !ok {
			continue
		}
		if err := t.Traverse(ft, append(fieldPath, field.Name())...); err != nil {
			return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
		}
	}
	return nil
}
