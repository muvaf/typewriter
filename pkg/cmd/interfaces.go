package cmd

import (
	"go/types"

	"github.com/muvaf/typewriter/pkg/packages"
)

type NewFuncGeneratorFn func(*packages.Cache, *packages.Imports) FuncGenerator

type FuncGenerator interface {
	Generate(t *types.Named, cm *packages.CommentMarkers) (map[string]interface{}, error)
	Matches(cm *packages.CommentMarkers) bool
}

type TypeGenerator interface {
	Generate() (*types.Named, *packages.CommentMarkers, error)
}
