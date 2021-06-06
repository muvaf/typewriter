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
	"go/types"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

type TypeGenerator interface {
	Generate() (*types.Named, *packages.CommentMarkers, error)
}

type NewFuncGeneratorFn func(*packages.Cache, *packages.Imports) FuncGenerator

type FuncGenerator interface {
	Generate(t *types.Named, cm *packages.CommentMarkers) (map[string]interface{}, error)
}

type FuncGeneratorChain []FuncGenerator

func (gc FuncGeneratorChain) Generate(t *types.Named, cm *packages.CommentMarkers) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for i, g := range gc {
		out, err := g.Generate(t, cm)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot run generator at index %d", i)
		}
		for k, v := range out {
			result[k] = v
		}
	}
	return result, nil
}
