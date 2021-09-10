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

package types

import (
	"fmt"
	"go/types"
)

// Comments holds all comments that will be printed together with objects. We
// have to have this map separate from the types because go/types package doesn't
// include comments anywhere in its structs.
type Comments map[string]string

// AddTypeComment lets you add comment for the given type.
func (c Comments) AddTypeComment(t *types.TypeName, comm string) {
	c[QualifiedTypePath(t)] = comm
}

// AddFieldComment lets you add comment for the given field of the type.
func (c Comments) AddFieldComment(t *types.TypeName, f, comm string) {
	c[QualifiedFieldPath(t, f)] = comm
}

// QualifiedTypePath returns a fully qualified path for the type name. The
// format is <package path>.<type name>
func QualifiedTypePath(n *types.TypeName) string {
	return fmt.Sprintf("%s.%s", n.Pkg().Path(), n.Name())
}

// QualifiedFieldPath returns a fully qualified path for the field of the given
// type. The format is <package path>.<type name>:<field name>
func QualifiedFieldPath(n *types.TypeName, f string) string {
	return fmt.Sprintf("%s.%s:%s", n.Pkg().Path(), n.Name(), f)
}
