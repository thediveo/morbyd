// Copyright 2026 Harald Albrecht.
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

package identity

import (
	"strconv"
	"strings"
)

// Principal is a string or int value to identify a specific user or group.
type Principal interface{ ~string | ~int }

// WithUser configures the user and optionally group the command(s) inside a
// container will be run as, taking either a user name, user:group names, or a
// user ID. WithUser will remove any previously configured group, either setting
// the specified group or configuring no group at all (so that the container
// image default applies).
func WithUser[I Principal](id I) string {
	var user string
	switch v := any(id).(type) {
	case string:
		user = v
	case int:
		user = strconv.FormatInt(int64(v), 10)
	}
	return user
}

// WithGroup configures the group the command(s) inside a container will be run
// as, taking either a group name or ID. If an empty group name "" is specified,
// any configured group name or ID will be removed.
func WithGroup[I Principal](p string, gid I) string {
	var group string
	switch v := any(gid).(type) {
	case string:
		group = v
	case int:
		group = strconv.FormatInt(int64(v), 10)
	}
	user, _, _ := strings.Cut(p, ":")
	if group == "" {
		return user
	}
	return user + ":" + group
}
