package cmd

import (
	"go/types"

	"github.com/muvaf/typewriter/pkg/packages"
)

type NewGeneratorFn func(*packages.Cache, *packages.Map) Generator

type Generator interface {
	Generate(t *types.Named, cm *packages.CommentMarkers) (map[string]interface{}, error)
	Matches(cm *packages.CommentMarkers) bool
}
