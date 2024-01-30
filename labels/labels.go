// Copyright 2024 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package labels

import (
	"fmt"
	"strings"
)

// Labels represents a set of labels (key-value pairs).
type Labels map[string]string

// MakeLabels parses the passed labels arguments in either form of “foo=bar” or
// “foo=”, returning a Labels map for further use in the Docker API. In case of
// an invalid label (such as "", "=", or "=bar"), MakeLabels returns an error.
func MakeLabels(labels ...string) (Labels, error) {
	l := Labels{}
	for _, label := range labels {
		if err := l.Add(label); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// Add a single label to a Labels map, where the label must be either in
// “foo=bar” or “foo=” form. If the specified label isn't in either form, or
// empty, an error is returned instead.
func (l Labels) Add(label string) error {
	key, value, ok := strings.Cut(label, "=")
	if !ok || key == "" {
		return fmt.Errorf("label must be in format \"KEY=\" or \"KEY=VALUE\", "+
			"got %q", label)
	}
	l[key] = value
	return nil
}
