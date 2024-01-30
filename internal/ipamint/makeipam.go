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
Package ipamint provides package “ipam”-internal APIs for sole use within the
morbyd module, not littering the public API.
*/
package ipamint

import "github.com/thediveo/morbyd/ipam"

// MakeIPAM returns a new IPAM configuration object, with the supplied
// IPAM-related options applied.
func MakeIPAM(opts ...ipam.IPAMOpt) (ipam.IPAM, error) {
	d := ipam.IPAM{}
	for _, opt := range opts {
		if err := opt(&d); err != nil {
			return ipam.IPAM{}, err
		}
	}
	return d, nil
}
