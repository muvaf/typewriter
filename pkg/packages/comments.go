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
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/tools/go/packages"
)

const (
	CommentPrefix = "+typewriter"
	SectionTypes  = "types"
	SectionMerged = "merged"
)

func NewCommentMarkers() *CommentMarkers {
	return &CommentMarkers{
		SectionContents: map[string][]string{},
	}
}

type CommentMarkers struct {
	SectionContents map[string][]string
}

func (ct *CommentMarkers) Print() string {
	out := ""
	for k, va := range ct.SectionContents {
		for _, v := range va {
			out += fmt.Sprintf("\n// +typewriter:%s:%s=%s", SectionTypes, k, v)
		}
	}
	return out
}

func FullPath(n *types.Named) string {
	return fmt.Sprintf("%s.%s", n.Obj().Pkg().Path(), n.Obj().Name())
}

func NewCommentTagFromText(c string) (*CommentMarkers, error) {
	if !strings.Contains(c, CommentPrefix) {
		return nil, nil
	}
	ct := NewCommentMarkers()
	lines := strings.Split(c, "\n")
	for _, l := range lines {
		if !strings.Contains(l, CommentPrefix) {
			continue
		}
		sections := strings.Split(l, ":")
		pair := strings.Split(sections[len(sections)-1], "=")
		if len(pair) > 2 {
			// TODO(muvaf): support multiple equalities in one line.
			return nil, errors.Errorf("there cannot be more than one equality sign in the marker")
		}
		k := pair[0]
		v := ""
		if len(pair) == 2 {
			v = pair[1]
		}
		if sections[1] == SectionTypes {
			if _, ok := ct.SectionContents[k]; !ok {
				ct.SectionContents[k] = nil
			}
			if v != "" {
				ct.SectionContents[k] = append(ct.SectionContents[k], v)
			}
		} else {
			return nil, errors.Errorf("only types section is currently supported in markers")
		}
	}
	return ct, nil
}

type objid struct {
	Filename string
	Line     int
}

func LoadCommentMarkers(p *packages.Package) (map[*types.Named]*CommentMarkers, error) {
	result := map[*types.Named]*CommentMarkers{}
	objPositions := map[objid]types.Object{}
	for _, n := range p.Types.Scope().Names() {
		o := p.Types.Scope().Lookup(n)
		pos := p.Fset.Position(o.Pos())
		objPositions[objid{Filename: pos.Filename, Line: pos.Line}] = o
	}
	for _, f := range p.Syntax {
		for _, g := range f.Comments {
			ct, err := NewCommentTagFromText(g.Text())
			if err != nil {
				return nil, err
			}
			if ct == nil {
				continue
			}
			pos := p.Fset.Position(g.End())
			belonging, ok := objPositions[objid{Filename: pos.Filename, Line: pos.Line + 1}]
			if !ok {
				continue
			}
			n, ok := belonging.Type().(*types.Named)
			if !ok {
				continue
			}
			result[n] = ct
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}
