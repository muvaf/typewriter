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

package packages

import (
	"fmt"
	"strings"
)

const (
	CommentPrefix = "+typewriter"
)

func NewCommentMarkers(c string) CommentMarkers {
	return CommentMarkers{
		Comment:         c,
		SectionContents: map[string]map[string]string{},
	}
}

type CommentMarkers struct {
	// SectionContents holds the equality pairs and indexed by the string until
	// the last ":".
	// For example, the following two lines:
	// +typewriter:types:key1=val1
	// +typewriter:types:key2=val2
	// would be indexed as following:
	// {
	//    "types": {"key1":"val1", "key2":"val2"}
	// }
	SectionContents map[string]map[string]string

	// Comment is the original comment string.
	Comment string
}

func (ct CommentMarkers) Print(prefix string) string {
	out := ""
	for section, pairs := range ct.SectionContents {
		for k, v := range pairs {
			out += fmt.Sprintf("\n// +%s:%s:%s=%s", prefix, section, k, v)
		}
	}
	return out
}

func NewCommentMarkersFromText(c string, prefix string) CommentMarkers {
	if !strings.Contains(c, prefix) {
		return CommentMarkers{}
	}
	ct := NewCommentMarkers(c)
	lines := strings.Split(c, "\n")
	for _, l := range lines {
		if !strings.Contains(l, prefix) {
			continue
		}
		l = strings.TrimPrefix(l, prefix)
		sections := strings.Split(l, ":")
		sectionKey := strings.Join(sections[:len(sections)-1], ":")
		pairs := strings.Split(sections[len(sections)-1], "=")
		if len(ct.SectionContents[sectionKey]) == 0 {
			ct.SectionContents[sectionKey] = map[string]string{}
		}
		val := ""
		if len(pairs) > 1 {
			val = pairs[1]
		}
		ct.SectionContents[sectionKey][pairs[0]] = val
	}
	return ct
}
