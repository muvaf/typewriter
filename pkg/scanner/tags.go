package scanner

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/tools/go/packages"
)

const (
	CommentPrefix     = "+typewriter"
	SectionTypes      = "types"
	SectionAggregated = "aggregated"
)

func NewCommentMarkers() *CommentMarkers {
	return &CommentMarkers{
		Types: map[string][]string{},
	}
}

type CommentMarkers struct {
	Types map[string][]string
}

func (ct *CommentMarkers) Print() string {
	out := ""
	for k, va := range ct.Types {
		for _, v := range va {
			out += fmt.Sprintf("// +typewriter:%s:%s=%s\n", SectionTypes, k, v)
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
			if _, ok := ct.Types[k]; !ok {
				ct.Types[k] = nil
			}
			if v != "" {
				ct.Types[k] = append(ct.Types[k], v)
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
