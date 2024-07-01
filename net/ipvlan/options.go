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

package ipvlan

import (
	"github.com/docker/docker/api/types/network"
	"github.com/thediveo/morbyd/internal/netint"
	"github.com/thediveo/morbyd/net"
)

type IPVLANMode string

const (
	L2Mode  IPVLANMode = "l2"
	L3Mode  IPVLANMode = "l3"
	L3sMode IPVLANMode = "l3s"
)

type IPVLANFlag string

const (
	BridgeFlag  IPVLANFlag = "bridge"
	VEPAFlag    IPVLANFlag = "vepa"
	PrivateFlag IPVLANFlag = "private"
)

// WithParent sets the name of the “parent” interface (and optional VLAN) to
// attach the IPVLAN “slave” interfaces to. In Linux parlance, the “parent”
// interface is commonly referred to as the “master”.
//
// See also: Docker's [IPvlan network driver options].
//
// [IPvlan network driver options]: https://docs.docker.com/network/drivers/ipvlan/#options
func WithParent(ifname string) net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*network.CreateOptions)(o))
		o.Options["parent"] = ifname
		return nil
	}
}

// WithMode sets the IPVLAN mode. If left unset, it defaults to “l2” [L2Mode].
// All slave “ipvlan” network interfaces will operate in the same mode.
//
//   - [L2Mode] – the parent network interfaces bridges L2 traffic between the
//     “slaves”, ARP can be used. However, different IP subnets on the same parent
//     network interface cannot ping each other. An external host only sees the
//     single external-facing MAC address of the parent.
//   - [L3Mode] – only IP-layer connectivity between “slaves”, where the parent
//     network interface acts as an IP router for unicast traffic. Different IP
//     subnets on the same parent network interface can reach each other using
//     IP unicast traffic.
//   - [L3sMode] – like L3Mode, but with packet filtering supported.
//
// See also:
//   - Docker's [IPvlan network driver options].
//   - [Introduction to Linux interfaces for virtual networking: IPVLAN].
//   - [Getting started with IPVLAN].
//   - Linux kernel [IPVLAN Driver HOWTO].
//
// [IPvlan network driver options]: https://docs.docker.com/network/drivers/ipvlan/#options
// [Introduction to Linux interfaces for virtual networking: IPVLAN]: https://developers.redhat.com/blog/2018/10/22/introduction-to-linux-interfaces-for-virtual-networking#ipvlan
// [IPVLAN Driver HOWTO]: https://docs.kernel.org/networking/ipvlan.html
// [Getting started with IPVLAN]: https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/system_design_guide/getting-started-with-ipvlan_system-design-guide
func WithMode(mode IPVLANMode) net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*network.CreateOptions)(o))
		o.Options["ipvlan_mode"] = string(mode)
		return nil
	}
}

// WithFlag sets the IPVLAN (mode) flag; when unset, it defaults to “bridge”
// (BridgeFlag).
//
//   - [BridgeFlag]
//   - [VEPAFlag]
//   - [PrivateFlag]
//
// See also: Docker's [IPvlan network driver options].
//
// [IPvlan network driver options]: https://docs.docker.com/network/drivers/ipvlan/#options
func WithFlag(flag IPVLANFlag) net.Opt {
	return func(o *net.Options) error {
		netint.EnsureOptionsMap((*network.CreateOptions)(o))
		o.Options["ipvlan_flag"] = string(flag)
		return nil
	}
}
