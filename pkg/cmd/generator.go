package cmd

import (
	"go/types"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

type GeneratorChain []Generator

func (gc GeneratorChain) Generate(t *types.Named, cm *packages.CommentMarkers) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for i, g := range gc {
		if !g.Matches(cm) {
			continue
		}
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

type Functions struct {
	Imports           *packages.Map
	SourcePackagePath string
	NewGeneratorFns   []NewGeneratorFn
	Cache             *packages.Cache
}

func (f *Functions) Run() (map[string]interface{}, error) {
	sourcePkg, err := f.Cache.GetPackage(f.SourcePackagePath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get source package")
	}
	recipe, err := packages.LoadCommentMarkers(sourcePkg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot scan comment markers")
	}
	gens := GeneratorChain{}
	for _, fn := range f.NewGeneratorFns {
		gens = append(gens, fn(f.Cache, f.Imports))
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
