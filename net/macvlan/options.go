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

package macvlan

import (
	"github.com/docker/docker/api/types"
	"github.com/thediveo/morbyd/internal/netint"
	"github.com/thediveo/morbyd/net"
)

type MACVLANMode string

const (
	BridgeMode   MACVLANMode = "bridge"
	VEPAMode     MACVLANMode = "vepa"
	PassthruMode MACVLANMode = "passthru"
	PrivateMode  MACVLANMode = "private"
)

// WithParent sets the name of the “parent” interface (and optional VLAN) to
// attach the MACVLAN “slave” interfaces to. In Linux parlance, the “parent”
// interface is commonly referred to as the “master”.
//
// See also: Docker's [Macvlan network driver options].
//
// [Macvlan network driver options]: https://docs.docker.com/network/drivers/macvlan/#options
func WithParent(ifname string) net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*types.NetworkCreate)(o))
		o.Options["parent"] = ifname
		return nil
	}
}

// WithMode sets the MACVLAN mode; when unset, it defaults to “bridge”
// (=BridgeMode).
//
//   - [BridgeMode] is the default mode and allows the MACVLANs of the
//     same “master” network interface to communicate other, but blocks
//     all traffic between the host and one of its MACVLAN network
//     interfaces.
//   - [VEPAMode] forwards all MACVLAN traffic to an external switch
//     that must support hair-pinning; that is, the external switch
//     must then forward traffic to other MACVLAN network interfaces back.
//   - [PassthruMode] is only allowed for one endpoint on the same master
//     network interface. It forces the master network interface into
//     promiscuous mode in order to bridge it or create VLAN network
//     interfaces on top of it.
//   - [PrivateMode] doesn't allow any traffic between MACVLANs on
//     the same master network interface.
//
// See also: Docker's [Macvlan network driver options].
//
// [Macvlan network driver options]: https://docs.docker.com/network/drivers/macvlan/#options
func WithMode(mode MACVLANMode) net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*types.NetworkCreate)(o))
		o.Options["macvlan_mode"] = string(mode)
		return nil
	}
}
