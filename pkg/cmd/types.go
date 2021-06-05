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
