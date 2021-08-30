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

// TODO(muvaf): Type can generate & print single type right now. However, that
// causes *types.Scope to be exposed as high level argument because we need to
// check for duplicates in the local package. If we accept only package path, then
// the package will be loaded (expensive) every time a type is generated. Consider
// the ability to allow multiple types to be printed.

func WithFlattenerOption(fo types.FlattenerOption) TypeOption {
	return func(t *Type) {
		t.FlattenerOption = fo
	}
}

type TypeOption func(*Type)

func NewType(im *packages.Imports, cache *packages.Cache, gen TypeGenerator, opts ...TypeOption) *Type {
	t := &Type{
		Imports:   im,
		Cache:     cache,
		Generator: gen,
	}
	for _, opt := range opts {
		opt(t)
	}

	return t
}

type Type struct {
	Imports         *packages.Imports
	Cache           *packages.Cache
	Generator       TypeGenerator
	FlattenerOption types.FlattenerOption
}

func (t *Type) Run() (string, error) {
	generated, _, err := t.Generator.Generate()
	if err != nil {
		return "", errors.Wrap(err, "cannot generate type")
	}
	flattened := types.NewFlattener(t.Imports, t.FlattenerOption,
		types.WithLocalPkg(generated.Obj().Pkg()),
	).Flatten(generated)
	printer := types.NewTypePrinter(t.Imports, generated.Obj().Pkg().Scope())
	structStr, err := printer.Print(flattened)
	return structStr, errors.Wrapf(err, "cannot print generated type %s", structStr)
}
