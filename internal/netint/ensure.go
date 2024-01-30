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

/*
Package netint provides package “net”-internal APIs for sole use within the
morbyd module, not littering the public API.
*/
package netint

import "github.com/docker/docker/api/types"

// EnsureLabelsMap is a helper to ensure that the net.Options.Labels map is
// initialized.
func EnsureLabelsMap(o *types.NetworkCreate) {
	if o.Labels != nil {
		return
	}
	o.Labels = map[string]string{}
}

// EnsureOptionsMap is a helper to ensure that the net.Options.Options map is
// initialized.
func EnsureOptionsMap(o *types.NetworkCreate) {
	if o.Options != nil {
		return
	}
	o.Options = map[string]string{}
}
