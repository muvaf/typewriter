package cmd

import (
	"go/token"
	"go/types"
	"strings"

	"github.com/pkg/errors"

	"github.com/muvaf/typewriter/pkg/packages"
)

type TypeGeneration struct {
	APIVersion string
	Kind       string
	Name       string
	Spec       TypeGenerationSpec

	imports *packages.Imports
	cache   *packages.Cache
}
type TypeGenerationSpec struct {
	Types []Type
}

type Type struct {
	PackagePath   string // optional
	TypeName      string
	GeneratorKind string // Aggregation
	Aggregation   *AggregationSpec
}

type AggregationSpec []TypeSignature

type TypeSignature struct {
	Package string
	Name    string
}

func (t *Type) Generate() (*types.Named, *packages.CommentMarkers, error) {
	varMap := map[string]*types.Var{}
	var cm *packages.CommentMarkers
	switch t.GeneratorKind {
	case "Aggregation":
		ag := AggregationGenerator{Spec: *t.Aggregation, cache: t}
	}
	name := types.NewTypeName(token.NoPos,
		types.NewPackage(t.PackagePath, t.PackagePath[strings.LastIndex(t.PackagePath, "/")+1:]),
		t.TypeName,
		nil)
	fields := make([]*types.Var, len(varMap))
	i := 0
	for _, v := range varMap {
		fields[i] = v
		i++
	}
	nn := types.NewNamed(name, types.NewStruct(fields, nil), nil)
	return nn, cm, nil
}

type AggregationGenerator struct {
	Spec  AggregationSpec
	cache *packages.Cache
}

func (ta *AggregationGenerator) Generate() (map[string]*types.Var, *packages.CommentMarkers, error) {
	varMap := map[string]*types.Var{}
	cm := packages.NewCommentMarkers()
	for _, call := range ta.Spec {
		pkg, err := ta.cache.GetPackage(call.Package)
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot get package")
		}
		t := pkg.Types.Scope().Lookup(call.Name)
		if t == nil {
			return nil, nil, errors.Errorf("cannot find type %s in package %s", call.Name, pkg.PkgPath)
		}
		cn := t.Type().(*types.Named)
		addAggregatedTypeMarker(cm, cn)
		str := cn.Underlying().(*types.Struct)
		for i := 0; i < str.NumFields(); i++ {
			varMap[str.Field(i).Name()] = str.Field(i)
		}
	}
	return varMap, cm, nil
}

func addAggregatedTypeMarker(cm *packages.CommentMarkers, n *types.Named) {
	fullPath := packages.FullPath(n)
	for _, ag := range cm.SectionContents[packages.SectionAggregated] {
		if ag == fullPath {
			return
		}
	}
	cm.SectionContents[packages.SectionAggregated] = append(cm.SectionContents[packages.SectionAggregated], fullPath)
}
