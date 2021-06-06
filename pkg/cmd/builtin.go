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
)

func NewProducers(cache *packages.Cache, im *packages.Imports) FuncGenerator {
	return &Producers{
		cache:   cache,
		imports: im,
	}
}

// Producers generates a function for every merged type of the given type that will
// let you produce those remote types from the local one.
type Producers struct {
	cache   *packages.Cache
	imports *packages.Imports
}

func (p *Producers) Generate(source *types.Named, cm *packages.CommentMarkers) (map[string]interface{}, error) {
	merged := cm.SectionContents[packages.SectionMerged]
	if len(merged) == 0 {
		return nil, nil
	}
	result := ""
	for _, target := range merged {
		targetType, err := p.cache.GetTypeWithFullPath(target)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get target type")
		}
		fn := traverser.NewPrinter(p.imports, traverser.NewGeneric(p.imports))
		funcName := fmt.Sprintf("Generate%s", targetType.Obj().Name())
		generated, err := fn.Print(funcName, source, targetType, nil)
		if err != nil {
			return nil, errors.Wrap(err, "cannot wrap function")
		}
		result += fmt.Sprintf("%s\n", generated)
	}
	return map[string]interface{}{
		"Producers": result,
	}, nil
}
