package traverser

import "go/types"

type GenericTraverser interface {
	Print(a, b types.Type, aFieldPath, bFieldPath string, levelNum int) (string, error)
}

type GenericCaller interface {
	SetGenericTraverser(p GenericTraverser)
}

type Templater interface {
	SetTemplate(t string)
}

type NamedTraverser interface {
	GenericCaller
	Print(a, b *types.Named, aFieldPath, bFieldPath string, levelNum int) (string, error)
}

type SliceTraverser interface {
	GenericCaller
	Templater
	Print(a, b *types.Slice, aFieldPath, bFieldPath string, levelNum int) (string, error)
}

type MapTraverser interface {
	GenericCaller
	Templater
	Print(a, b *types.Map, aFieldPath, bFieldPath string, levelNum int) (string, error)
}

type PointerTraverser interface {
	GenericCaller
	Templater
	Print(a, b *types.Pointer, aFieldPath, bFieldPath string, levelNum int) (string, error)
}

type BasicTraverser interface {
	SetTemplate(t map[types.BasicKind]string)
	Print(a, b *types.Basic, aFieldPath, bFieldPath string) (string, error)
}