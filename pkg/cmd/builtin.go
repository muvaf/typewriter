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

package cmd

import (
	"fmt"
	"go/types"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
	"github.com/muvaf/typewriter/pkg/traverser"
	"github.com/muvaf/typewriter/pkg/wrapper"
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
	aggregated := cm.SectionContents[packages.SectionMerged]
	for _, target := range aggregated {
		targetType, err := p.cache.GetTypeWithFullPath(target)
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
	return len(cm.SectionContents[packages.SectionMerged]) > 0
}
