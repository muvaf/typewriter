package cmd

import (
	"fmt"
	"go/types"

	"github.com/muvaf/typewriter/pkg/imports"
	"github.com/muvaf/typewriter/pkg/traverser"
	"github.com/muvaf/typewriter/pkg/wrapper"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/scanner"
)

func NewProducer(cache *Cache, im *imports.Map) Generator {
	if cache == nil {
		cache = NewCache()
	}
	return &Producer{
		cache:   cache,
		imports: im,
	}
}

type Producer struct {
	cache   *Cache
	imports *imports.Map
}

func (p *Producer) Generate(source *types.Named, cm *scanner.CommentMarkers) (map[string]interface{}, error) {
	result := ""
	aggregated := cm.Types[scanner.SectionAggregated]
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

func (p *Producer) Matches(cm *scanner.CommentMarkers) bool {
	_, ok := cm.Types[scanner.SectionAggregated]
	return ok && len(cm.Types[scanner.SectionAggregated]) > 0
}
