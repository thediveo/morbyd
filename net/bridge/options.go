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

package bridge

import (
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/thediveo/morbyd/internal/netint"
	"github.com/thediveo/morbyd/net"
)

const (
	networkOptionStem = "com.docker.network."
	bridgeOptionStem  = networkOptionStem + "bridge."
	driverOptionStem  = networkOptionStem + "driver."
)

// WithBridgeName instructs the “bridge” driver to use a pre-created or create a
// Linux kernel bridge with the specified name. If left unset, the “bridge”
// driver will automatically create a Linux kernel bridge with a name in
// “br-XXXXXXXXXXXX” format, where XXXXXXXXXXXX are taken from the first 12 hex
// digits of the Docker network ID.
//
// See also: Docker's [Bridge network driver options].
//
// [Bridge network driver options]: https://docs.docker.com/network/drivers/bridge/#options
func WithBridgeName(ifname string) net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*types.NetworkCreate)(o))
		o.Options[bridgeOptionStem+"name"] = ifname
		return nil
	}
}

// WithoutIPMasquerade disables IP masquerading for this custom Docker network.
//
// See also: Docker's [Bridge network driver options].
//
// [Bridge network driver options]: https://docs.docker.com/network/drivers/bridge/#options
func WithoutIPMasquerade() net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*types.NetworkCreate)(o))
		o.Options[bridgeOptionStem+"enable_ip_masquerade"] = "false"
		return nil
	}
}

// WithoutICC disabled inter-container communication on this custom Docker
// network.
//
// See also: Docker's [Bridge network driver options].
//
// [Bridge network driver options]: https://docs.docker.com/network/drivers/bridge/#options
func WithoutICC() net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*types.NetworkCreate)(o))
		o.Options[bridgeOptionStem+"enable_icc"] = "false"
		return nil
	}
}

// WithMTU sets the MTU size.
//
// See also: Docker's [Bridge network driver options].
//
// [Bridge network driver options]: https://docs.docker.com/network/drivers/bridge/#options
func WithMTU(mtu uint) net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*types.NetworkCreate)(o))
		o.Options[driverOptionStem+"mtu"] = strconv.FormatUint(uint64(mtu), 10)
		return nil
	}
}

// WithInterfacePrefix sets the prefix to use for names of the container network
// interfaces attached to this custom Docker network. Defaults to “eth” prefix
// when left unset.
//
// See also: Docker's [Bridge network driver options].
//
// [Bridge network driver options]: https://docs.docker.com/network/drivers/bridge/#options
func WithInterfacePrefix(prefix string) net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*types.NetworkCreate)(o))
		o.Options[networkOptionStem+"container_iface_prefix"] = prefix
		return nil
	}
}
