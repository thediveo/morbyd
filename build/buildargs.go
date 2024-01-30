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

package build

import (
	"fmt"
	"strings"
)

// BuildArgs represents (image) build arguments, where these build arguments can
// either represent key-value pairs (including empty values), or a nil-valued
// key. A nil-valued key indicates that the default value should be taken (as
// opposed to an empty “” value).
type BuildArgs map[string]*string

// MakeBuildArgs parses the passed build arguments in either of the three forms
// “foo=bar”, “foo=” and “foo”, returning a BuildArgs map for use in building
// image options. In case of invalid build arguments, such as “=” and “=foo”, it
// returns an error instead of nil.
func MakeBuildArgs(bargs ...string) (BuildArgs, error) {
	b := BuildArgs{}
	for _, barg := range bargs {
		if err := b.Add(barg); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// Add a single build argument to a BuildArgs map, where the build argument is
// in either of the three forms “foo=bar”, “foo=” and “foo”. If the build
// argument is in an invalid format, then Add returns an error, otherwise nil.
func (b BuildArgs) Add(barg string) error {
	key, value, ok := strings.Cut(barg, "=")
	if key == "" {
		return fmt.Errorf("invalid build arg format, expected \"KEY=VALUE\", \"KEY=\", or \"KEY\", got %q",
			barg)
	}
	if !ok {
		b[key] = nil
		return nil
	}
	b[key] = &value
	return nil
}
