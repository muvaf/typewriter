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

func WithFieldProcessor(fp FieldProcessor) TraverserOption {
	return func(t *Traverser) {
		t.FieldProcessors = fp
	}
}

func WithTypeProcessor(tp TypeProcessor) TraverserOption {
	return func(t *Traverser) {
		t.TypeProcessors = tp
	}
}

type TraverserOption func(*Traverser)

func NewTraverser(cache *packages.Cache, opts ...TraverserOption) *Traverser {
	t := &Traverser{
		cache:           cache,
		commentCache:    packages.NewCommentCache(cache),
		TypeProcessors:  TypeProcessorChain{},
		FieldProcessors: FieldProcessorChain{},
	}
	for _, f := range opts {
		f(t)
	}
	return t
}

type Traverser struct {
	TypeProcessors  TypeProcessor
	FieldProcessors FieldProcessor

	cache        *packages.Cache
	commentCache *packages.CommentCache
}

func (t *Traverser) Traverse(n *types.Named, formerFields ...string) error {
	pComments, err := t.commentCache.GetPackageComments(n.Obj().Pkg().Path())
	if err != nil {
		return errors.Wrapf(err, "cannot get comments for package %s", n.Obj().Pkg().Path())
	}
	if err := t.TypeProcessors.Process(n, pComments.CommentOf(n.Obj())); err != nil {
		return errors.Wrapf(err, "type processors failed to run for type %s", n.Obj().Name())
	}
	st, ok := n.Underlying().(*types.Struct)
	if !ok {
		return nil
	}
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		tag := st.Tag(i)
		if err := t.FieldProcessors.Process(n, field, tag, pComments.CommentOf(field), formerFields); err != nil {
			return errors.Wrapf(err, "field processors failed to run for field %s of type %s", field.Name(), n.Obj().Name())
		}
		switch ft := field.Type().(type) {
		case *types.Named:
			if err := t.Traverse(ft, append(formerFields, field.Name())...); err != nil {
				return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
			}
		case *types.Pointer:
			switch elemType := ft.Elem().(type) {
			case *types.Named:
				if err := t.Traverse(elemType, append(formerFields, "*", field.Name())...); err != nil {
					return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
				}
			}
		case *types.Slice:
			switch elemType := ft.Elem().(type) {
			case *types.Named:
				if err := t.Traverse(elemType, append(formerFields, "[]", field.Name())...); err != nil {
					return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
				}
			case *types.Pointer:
				switch elemElemType := elemType.Elem().(type) {
				case *types.Named:
					if err := t.Traverse(elemElemType, append(formerFields, "[]", "*", field.Name())...); err != nil {
						return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
					}
				}
			}
		case *types.Map:
			switch elemType := ft.Elem().(type) {
			case *types.Named:
				if err := t.Traverse(elemType, append(formerFields, "[mapvalue]", field.Name())...); err != nil {
					return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
				}
			case *types.Pointer:
				switch elemElemType := elemType.Elem().(type) {
				case *types.Named:
					if err := t.Traverse(elemElemType, append(formerFields, "[mapvalue]", "*", field.Name())...); err != nil {
						return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
					}
				}
			}
			switch keyType := ft.Key().(type) {
			case *types.Named:
				if err := t.Traverse(keyType, append(formerFields, "[mapkey]", field.Name())...); err != nil {
					return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
				}
			case *types.Pointer:
				switch elemKeyType := keyType.Elem().(type) {
				case *types.Named:
					if err := t.Traverse(elemKeyType, append(formerFields, "[mapkey]", "*", field.Name())...); err != nil {
						return errors.Wrapf(err, "failed to traverse type of field %s", field.Name())
					}
				}
			}
		}
	}
	return nil
}
