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

package ipam

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/network"
)

// IPAMOpt is a configuration option for configuring an IPAM driver (or “IPAM”
// for short).
type IPAMOpt func(*IPAM) error

// IPAM represents Docker's IPAM driver options, including IP address pool
// configuration, for creating a new custom Docker network.
type IPAM network.IPAM

// WithName sets the name of the IPAM driver to use. When unset, it defaults to
// “default” (sic!), see also [DefaultIPAM driver name] in the Moby libnetwork
// codebase.
//
// Please note that the “null” IPAM driver cannot be used with at least some of
// the network drivers, as these explicitly reject any attempt to use the “null”
// IPAM in order to not automatically assign any container IP addresses. A
// prominent example is the “macvlan” network driver.
//
// [DefaultIPAM driver name]: https://github.com/moby/libnetwork/blob/3797618f9a38372e8107d8c06f6ae199e1133ae8/ipamapi/contract.go#L18
func WithName(name string) IPAMOpt {
	return func(d *IPAM) error {
		d.Driver = name
		return nil
	}
}

// WithPool adds an IP address pool for the specified subnet, and accepting
// further pool configuration options (see [PoolOpt]).
func WithPool(subnet string, opts ...PoolOpt) IPAMOpt {
	return func(d *IPAM) error {
		pool, err := makePool(subnet, opts...)
		if err != nil {
			return err
		}
		d.Config = append(d.Config, network.IPAMConfig(pool))
		return nil
	}
}

// WithOption specifies a IPAM driver-specific option in “KEY=VALUE” format.
func WithOption(option string) IPAMOpt {
	return func(d *IPAM) error {
		key, value, ok := strings.Cut(option, "=")
		if !ok || key == "" {
			return fmt.Errorf("IPAM driver option must be in format \"KEY=\" or \"KEY=VALUE\", "+
				"got %q", option)
		}
		if d.Options == nil {
			d.Options = map[string]string{}
		}
		d.Options[key] = value
		return nil
	}
}
