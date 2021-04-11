package cmd

import (
	"fmt"
	"go/types"

	"github.com/muvaf/typewriter/pkg/packages"
	"github.com/muvaf/typewriter/pkg/traverser"
	"github.com/muvaf/typewriter/pkg/wrapper"

	"github.com/pkg/errors"
)

func NewProducer(cache *packages.Cache, im *packages.Imports) FuncGenerator {
	if cache == nil {
		cache = packages.NewCache()
	}
	return &Producer{
		cache:   cache,
		imports: im,
	}
}

type Producer struct {
	cache   *packages.Cache
	imports *packages.Imports
}

func (p *Producer) Generate(source *types.Named, cm *packages.CommentMarkers) (map[string]interface{}, error) {
	result := ""
	aggregated := cm.SectionContents[packages.SectionAggregated]
	for _, target := range aggregated {
		targetType, err := p.cache.GetType(target)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get target type")
		}
		fn := wrapper.NewFunc(p.imports, traverser.NewGeneric(p.imports))
		generated, err := fn.Wrap(fmt.Sprintf("Generate%s", targetType.Obj().Name()), source, targetType, nil)
		if err != nil {
			return nil, errors.Wrap(err, "cannot wrap function")
		}
		result += fmt.Sprintf("%s\n", generated)
	}
	return map[string]interface{}{
		"Producers": result,
	}, nil
}

func (p *Producer) Matches(cm *packages.CommentMarkers) bool {
	return len(cm.SectionContents[packages.SectionAggregated]) > 0
}
