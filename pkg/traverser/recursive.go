package typewriter

import (
	"fmt"
	"github.com/muvaf/typewriter/pkg/imports"
	"github.com/pkg/errors"
	"go/types"
)

type TypeTraverser interface {
	Print(a, b types.Type, aFieldPath, bFieldPath string, levelNum int) (string, error)
}

type RecursiveCaller interface {
	SetTypeTraverser(p TypeTraverser)
}

type NamedTraverser interface {
	RecursiveCaller
	Print(a, b *types.Named, aFieldPath, bFieldPath string, levelNum int) (string, error)
}

type SliceTraverser interface {
	RecursiveCaller
	Print(a, b *types.Slice, aFieldPath, bFieldPath string, levelNum int) (string, error)
}

type BasicTraverser interface {
	Print(a, b *types.Basic, aFieldPath, bFieldPath string) (string, error)
}

func WithBasic(p BasicTraverser) Option {
	return func(r *Type) {
		r.Basic = p
	}
}

func WithNamed(p NamedTraverser) Option {
	return func(r *Type) {
		p.SetTypeTraverser(r)
		r.Named = p
	}
}

func WithSlice(p SliceTraverser) Option {
	return func(r *Type) {
		p.SetTypeTraverser(r)
		r.Slice = p
	}
}

type Option func(*Type)

func NewType(im imports.Map, opts ...Option) *Type {
	r := &Type{
		Imports: im,
		Slice: NewSlice(),
		Named: NewNamed(),
		Basic: NewBasic(),
	}
	for _, f := range opts {
		f(r)
	}
	r.Slice.SetTypeTraverser(r)
	r.Named.SetTypeTraverser(r)
	return r
}

type Type struct {
	Imports imports.Map
	Named NamedTraverser
	Slice SliceTraverser
	Basic BasicTraverser
}

func (r *Type) Print(a, b types.Type, aFieldPath, bFieldPath string, levelNum int) (string, error) {
	switch at := a.(type) {
	case *types.Pointer:
		bt, ok := b.(*types.Pointer)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		o, err := r.Print(at.Elem(), bt.Elem(), aFieldPath, bFieldPath, levelNum)
		return o, errors.Wrap(err, "cannot recursively traverse actual type of pointer type")
	case *types.Slice:
		bt, ok := b.(*types.Slice)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		o, err := r.Slice.Print(at, bt, aFieldPath, bFieldPath, levelNum)
		return o, errors.Wrap(err, "cannot traverse slice type")
	case *types.Named:
		bt, ok := b.(*types.Named)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		o, err := r.Named.Print(at, bt, aFieldPath, bFieldPath, levelNum)
		return o, errors.Wrap(err, "cannot traverse named type")
	case *types.Basic:
		bt, ok := b.(*types.Basic)
		if !ok {
			return "", fmt.Errorf("not same type at %s", bFieldPath)
		}
		o, err := r.Basic.Print(at, bt, aFieldPath, bFieldPath)
		return o, errors.Wrap(err, "cannot traverse basic type")
	case *types.Struct: // unnamed struct fields.
		return "", nil
	default:
		return "", fmt.Errorf("unknown type in recursion: %s\n", at.String())
	}
}
