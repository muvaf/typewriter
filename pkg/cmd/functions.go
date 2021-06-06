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
	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

func WithNewFuncGeneratorFns(fn ...NewFuncGeneratorFn) FunctionsOption {
	return func(f *Functions) {
		f.NewGeneratorFns = fn
	}
}

type FunctionsOption func(*Functions)

func NewFunctions(c *packages.Cache, i *packages.Imports, sourcePkgPath string, opts ...FunctionsOption) *Functions {
	f := &Functions{
		cache:             c,
		imports:           i,
		SourcePackagePath: sourcePkgPath,
	}

	for _, opt := range opts {
		opt(f)
	}
	return f
}

type Functions struct {
	cache   *packages.Cache
	imports *packages.Imports

	SourcePackagePath string
	NewGeneratorFns   []NewFuncGeneratorFn
}

func (f *Functions) Run() (map[string]interface{}, error) {
	sourcePkg, err := f.cache.GetPackage(f.SourcePackagePath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get source package")
	}
	recipe, err := packages.LoadCommentMarkers(sourcePkg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot scan comment markers")
	}
	gens := FuncGeneratorChain{}
	for _, fn := range f.NewGeneratorFns {
		gens = append(gens, fn(f.cache, f.imports))
	}
	input := map[string]interface{}{}
	for sourceType, commentMarker := range recipe {
		generated, err := gens.Generate(sourceType, commentMarker)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot run generators for type %s", sourceType.Obj().Name())
		}
		for k, v := range generated {
			input[k] = v
		}
	}
	return input, nil
}
