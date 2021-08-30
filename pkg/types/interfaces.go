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
	"go/types"

	"github.com/pkg/errors"
)

type TypeProcessorChain []TypeProcessor

func (tpc TypeProcessorChain) Process(n *types.Named, comment string) error {
	for i, tp := range tpc {
		if err := tp.Process(n, comment); err != nil {
			return errors.Errorf("type processor at index %d failed", i)
		}
	}
	return nil
}

type TypeProcessor interface {
	Process(n *types.Named, comment string) error
}

type FieldProcessorChain []FieldProcessor

func (fpc FieldProcessorChain) Process(n *types.Named, f *types.Var, tag string, comment string, formerFields []string) error {
	for i, fp := range fpc {
		if err := fp.Process(n, f, tag, comment, formerFields); err != nil {
			return errors.Errorf("field processor at index %d failed", i)
		}
	}
	return nil
}

type FieldProcessor interface {
	Process(n *types.Named, f *types.Var, tag string, comment string, formerFields []string) error
}
