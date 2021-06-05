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
	"github.com/muvaf/typewriter/pkg/types"
)

func NewType(im *packages.Imports, cache *packages.Cache, gen TypeGenerator) *Type {
	return &Type{
		Imports:   im,
		Cache:     cache,
		Generator: gen,
	}
}

type Type struct {
	Imports   *packages.Imports
	Cache     *packages.Cache
	Generator TypeGenerator
}

func (t *Type) Run() (string, error) {
	result, markers, err := t.Generator.Generate()
	if err != nil {
		return "", errors.Wrap(err, "cannot generate type")
	}
	printer := types.NewTypePrinter(t.Imports, result.Obj().Pkg().Scope())
	structStr, err := printer.Print(result, markers.Print())
	if err != nil {
		return "", errors.Wrapf(err, "cannot print generated type %s", structStr)
	}
	return structStr, nil
}
