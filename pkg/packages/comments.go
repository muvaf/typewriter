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
	"go/token"
	"go/types"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

// Comments lets you fetch comment of an object.
type Comments struct {
	fset  *token.FileSet
	store map[objPos]string
}

type objPos struct {
	filename string
	line     int
}

func (c Comments) CommentOf(o types.Object) string {
	pos := c.fset.Position(o.Pos())
	return c.store[objPos{filename: pos.Filename, line: pos.Line}]
}

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

	// cache holds comments for all packages that have been loaded.
	cache map[string]Comments
}

func (cc *CommentCache) GetPackageComments(pkgPath string) (Comments, error) {
	p, err := cc.pkgCache.GetPackage(pkgPath)
	if err != nil {
		return Comments{}, errors.Wrapf(err, "cannot get package %s", pkgPath)
	}
	return LoadComments(p), nil
}

func LoadComments(p *packages.Package) Comments {
	result := Comments{
		fset:  p.Fset,
		store: map[objPos]string{},
	}
	for _, f := range p.Syntax {
		for _, c := range f.Comments {
			if len(c.Text()) == 0 {
				continue
			}
			pos := p.Fset.Position(c.End())
			result.store[objPos{filename: pos.Filename, line: pos.Line + 1}] = c.Text()
		}
	}
	return result
}
