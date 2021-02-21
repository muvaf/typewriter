package typewriter

import (
	"fmt"
	"go/types"
)

type TypeTraverser interface {
	Print(a, b types.Type, aFieldPath, bFieldPath string, levelNum int) string
}

type RecursiveCaller interface {
	SetTypeTraverser(p TypeTraverser)
}

type NamedTraverser interface {
	RecursiveCaller
	Print(a, b *types.Named, aFieldPath, bFieldPath string, levelNum int) string
}

type SliceTraverser interface {
	RecursiveCaller
	Print(a, b *types.Slice, aFieldPath, bFieldPath string, levelNum int) string
}

type BasicTraverser interface {
	Print(a, b *types.Basic, aFieldPath, bFieldPath string, levelNum int) string
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

func NewType(opts ...Option) *Type {
	r := &Type{
		Slice: NewSlice(),
		Named: NewNamed(),
		Basic: NewBasic(),
	}
	r.Slice.SetTypeTraverser(r)
	r.Named.SetTypeTraverser(r)
	for _, f := range opts {
		f(r)
	}
	return r
}

type Type struct {
	Named NamedTraverser
	Slice SliceTraverser
	Basic BasicTraverser
}

func (r *Type) Print(a, b types.Type, aFieldPath, bFieldPath string, levelNum int) string {
	switch at := a.(type) {
	case *types.Pointer:
		bt, ok := b.(*types.Pointer)
		if !ok {
			return fmt.Sprintf("not same type at %s", bFieldPath)
		}
		return r.Print(at.Elem(), bt.Elem(), aFieldPath, bFieldPath, levelNum)
	case *types.Slice:
		bt, ok := b.(*types.Slice)
		if !ok {
			return fmt.Sprintf("not same type at %s", bFieldPath)
		}
		return r.Slice.Print(at, bt, aFieldPath, bFieldPath, levelNum)
	case *types.Named:
		bt, ok := b.(*types.Named)
		if !ok {
			return fmt.Sprintf("not same type at %s", bFieldPath)
		}
		return r.Named.Print(at, bt, aFieldPath, bFieldPath, levelNum)
	case *types.Basic:
		bt, ok := b.(*types.Basic)
		if !ok {
			return fmt.Sprintf("not same type at %s", bFieldPath)
		}
		return r.Basic.Print(at, bt, aFieldPath, bFieldPath, levelNum)
	case *types.Struct: // unnamed struct fields.
		return ""
	default:
		return fmt.Sprintf("unknown type in recur: %s\n", at.String())
	}
}
