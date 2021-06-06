// Copyright 2021 Muvaffak Onus
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	SetPointerTemplate(t map[types.BasicKind]string)
	Print(a, b *types.Basic, aFieldPath, bFieldPath string, isPointer bool) (string, error)
}
