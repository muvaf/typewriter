package cmd

import (
	"go/types"

	"github.com/muvaf/typewriter/pkg/imports"
	"github.com/muvaf/typewriter/pkg/scanner"
)

type NewGeneratorFn func(*Cache, *imports.Map) Generator

type Generator interface {
	Generate(t *types.Named, cm *scanner.CommentMarkers) (map[string]interface{}, error)
	Matches(cm *scanner.CommentMarkers) bool
}
