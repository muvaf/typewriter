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

import "go/types"

type TypeFilterChain []TypeFilter

func (tc TypeFilterChain) Filter(t *types.Named) *types.Named {
	for _, tf := range tc {
		if t == nil {
			return nil
		}
		t = tf.Filter(t)
	}
	return t
}

type NopTypeFilter struct{}

func (tc NopTypeFilter) Filter(t *types.Named) *types.Named {
	return t
}

type FieldFilterChain []FieldFilter

func (tc FieldFilterChain) Filter(field *types.Var, tag string) (*types.Var, string) {
	for _, tf := range tc {
		if field == nil {
			return nil, ""
		}
		field, tag = tf.Filter(field, tag)
	}
	return field, tag
}

type NopFieldFilter struct{}

func (tc NopFieldFilter) Filter(field *types.Var, tag string) (*types.Var, string) {
	return field, tag
}

func NewIgnoreTypeFilter(list []string) *IgnoreTypeFilter {
	ig := &IgnoreTypeFilter{}
	for _, t := range list {
		ig.ignore[t] = struct{}{}
	}
	return ig
}

type IgnoreTypeFilter struct {
	ignore map[string]struct{}
}

func (ig *IgnoreTypeFilter) Filter(name types.TypeName, t types.Type) types.Type {
	if _, ok := ig.ignore[name.Name()]; ok {
		return nil
	}
	return t
}

func NewIgnoreFieldFilter(list []string) *IgnoreFieldFilter {
	ig := &IgnoreFieldFilter{}
	for _, t := range list {
		ig.ignore[t] = struct{}{}
	}
	return ig
}

type IgnoreFieldFilter struct {
	ignore map[string]struct{}
}

func (ig *IgnoreFieldFilter) Filter(field *types.Var, tag string) (*types.Var, string) {
	if _, ok := ig.ignore[field.Name()]; ok {
		return nil, ""
	}
	return field, tag
}
