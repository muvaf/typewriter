package typewriter

import (
	"fmt"
	"go/types"
)

type GeneralPrinter interface {
	Print(a, b types.Type, aFieldPath, bFieldPath string, levelNum int) string
}

type RecursiveCaller interface {
	SetGeneralPrinter(p GeneralPrinter)
}

type NamedPrinter interface {
	RecursiveCaller
	Print(a, b *types.Named, aFieldPath, bFieldPath string, levelNum int) string
}

type SlicePrinter interface {
	RecursiveCaller
	Print(a, b *types.Slice, aFieldPath, bFieldPath string, levelNum int) string
}

type BasicPrinter interface {
	Print(a, b *types.Basic, aFieldPath, bFieldPath string, levelNum int) string
}

func WithBasicPrinter(p BasicPrinter) Option {
	return func(r *RecursivePrinter) {
		r.Basic = p
	}
}

func WithNamedPrinter(p NamedPrinter) Option {
	return func(r *RecursivePrinter) {
		p.SetGeneralPrinter(r)
		r.Named = p
	}
}

func WithSlicePrinter(p SlicePrinter) Option {
	return func(r *RecursivePrinter) {
		p.SetGeneralPrinter(r)
		r.Slice = p
	}
}

type Option func(*RecursivePrinter)

func NewRecursivePrinter(opts ...Option) *RecursivePrinter {
	r := &RecursivePrinter{
		Slice: NewSlice(),
		Named: NewNamed(),
		Basic: NewBasic(),
	}
	r.Slice.SetGeneralPrinter(r)
	r.Named.SetGeneralPrinter(r)
	for _, f := range opts {
		f(r)
	}
	return r
}

type RecursivePrinter struct {
	Named  NamedPrinter
	Slice  SlicePrinter
	Basic  BasicPrinter
}

func (r *RecursivePrinter) Print(a, b types.Type, aFieldPath, bFieldPath string, levelNum int) string {
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
