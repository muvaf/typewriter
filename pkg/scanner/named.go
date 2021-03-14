package scanner

import (
	"fmt"
	"go/types"
)

const (
	AggregatedTypeCommentPrefixFmt = "// +typewriter:types:aggregated:%s"
)

func NewNamed(n *types.Named, ct []string) *Named {
	return &Named{
		Named:       n,
		CommentTags: ct,
	}
}

type Named struct {
	*types.Named
	CommentTags []string
}

func AggregatedTypesTags(types []string) []string {
	var result []string
	for _, t := range types {
		result = append(result, fmt.Sprintf(AggregatedTypeCommentPrefixFmt, t))
	}
	return result
}
