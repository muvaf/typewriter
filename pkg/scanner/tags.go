package scanner

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

func NewCommentTags() *CommentTags {
	return &CommentTags{
		types: typeTags{
			aggregated: map[string]struct{}{},
		},
	}
}

type CommentTags struct {
	types typeTags
}

type typeTags struct {
	aggregated map[string]struct{}
}

func (ct *CommentTags) AddAggregated(t string) {
	ct.types.aggregated[t] = struct{}{}
}

func (ct *CommentTags) GetAggregatedTypes() map[string]struct{} {
	return ct.types.aggregated
}

// TODO(muvaf): Support comma-seperated lists.

func (ct *CommentTags) Print() string {
	out := ""
	for a := range ct.types.aggregated {
		out += fmt.Sprintf("// +typewriter:types:aggregated=%s\n", a)
	}
	return out
}

func NewCommentTagFromText(c string) *CommentTags {
	// This is a temporary implementation.
	prefix := fmt.Sprintf("+typewriter:types:aggregated=")
	ct := NewCommentTags()
	lines := strings.Split(c, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			ct.AddAggregated(strings.TrimPrefix(l, prefix))
		}
	}
	if len(ct.types.aggregated) == 0 {
		return nil
	}
	return ct
}

type objid struct {
	Filename string
	Line     int
}

func ScanCommentTags(p *packages.Package) map[*types.Named]*CommentTags {
	result := map[*types.Named]*CommentTags{}
	objPositions := map[objid]types.Object{}
	for _, n := range p.Types.Scope().Names() {
		o := p.Types.Scope().Lookup(n)
		pos := p.Fset.Position(o.Pos())
		objPositions[objid{Filename: pos.Filename, Line: pos.Line}] = o
	}
	for _, f := range p.Syntax {
		for _, g := range f.Comments {
			ct := NewCommentTagFromText(g.Text())
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
		return nil
	}
	return result
}
