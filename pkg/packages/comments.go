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

package packages

import (
	"fmt"
	"go/types"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

// Comments lets you fetch comment of an object.
type Comments map[types.Object]string

func NewCommentCache(cache *Cache) *CommentCache {
	return &CommentCache{
		pkgCache: cache,
		cache:    map[string]Comments{},
	}
}

// CommentCache serves as the cache for accessing comments in packages. Indexed
// by package path.
type CommentCache struct {
	pkgCache *Cache
	cache    map[string]Comments
}

func (cc *CommentCache) GetComments(pkgPath string) (Comments, error) {
	p, err := cc.pkgCache.GetPackage(pkgPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get package %s", pkgPath)
	}
	c, err := LoadComments(p)
	return c, errors.Wrapf(err, "cannot load comments for package %s", pkgPath)
}

type objid struct {
	Filename string
	Line     int
}

func LoadComments(p *packages.Package) (Comments, error) {
	result := Comments{}
	objPositions := map[objid]types.Object{}
	for _, n := range p.Types.Scope().Names() {
		o := p.Types.Scope().Lookup(n)
		pos := p.Fset.Position(o.Pos())
		objPositions[objid{Filename: pos.Filename, Line: pos.Line}] = o
	}
	for _, f := range p.Syntax {
		for _, g := range f.Comments {
			if len(g.Text()) == 0 {
				continue
			}
			pos := p.Fset.Position(g.End())
			belonging, ok := objPositions[objid{Filename: pos.Filename, Line: pos.Line + 1}]
			if !ok {
				continue
			}
			result[belonging] = g.Text()
		}
	}
	return result, nil
}

func FullPath(n *types.Named) string {
	return fmt.Sprintf("%s.%s", n.Obj().Pkg().Path(), n.Obj().Name())
}
