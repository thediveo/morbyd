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

package morbyd

import (
	"math/rand"
	"net"
	"net/netip"
)

// Addr represents a non-zero IP address together with a transport-layer port
// and the particular transport-layer protocol (or “network” in Go parlance).
type Addr struct {
	addrport netip.AddrPort
	network  string
}

// Addrs is a list of Addr elements, providing helper functions on top. See
// [Addrs.Any] and [Addrs.First].
type Addrs []Addr

var _ (net.Addr) = (*Addr)(nil)

// NewAddr returns an Addr element, configured from the passed [netip.AddrPort]
// and transport-layer protocol, a.k.a. “network” (such as “tcp”).
func NewAddr(addrport netip.AddrPort, network string) Addr {
	return Addr{
		addrport: addrport,
		network:  network,
	}
}

// Network returns the “name” of the “network” the service is on, either “tcp”
// or “udp”.
func (a Addr) Network() string {
	return a.network
}

// String returns the address with port, such as “127.0.0.1:123” or “[::1]:123”.
// It returns "" for a zero Addr.
func (a Addr) String() string {
	if a.network == "" {
		return ""
	}
	return a.addrport.String()
}

// UnspecifiedAsLoopback returns the IPv4 or IPv6 loopback address if this Addr
// is unspecified, otherwise Addr unmodified. Use UnspecifiedAsLoopback in
// chained Addr operations to ensure to always get a non-unspecified destination
// address of a published container port.
//
//	addr := addrs.First().UnspecifiedAsLoopback()
func (a Addr) UnspecifiedAsLoopback() Addr {
	if a.network == "" {
		return Addr{}
	}
	if !a.addrport.Addr().IsUnspecified() {
		return a
	}
	if a.addrport.Addr().Is4() {
		return Addr{
			addrport: netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), a.addrport.Port()),
			network:  a.network,
		}
	}
	return Addr{
		addrport: netip.AddrPortFrom(netip.IPv6Loopback(), a.addrport.Port()),
		network:  a.network,
	}
}

// Any returns a random address element from this address list. If the address
// list is empty, it returns a zero address element instead.
func (a Addrs) Any() Addr {
	if len(a) == 0 {
		return Addr{}
	}
	return a[rand.Intn(len(a))]
}

// First always returns the first address element from this address list. If the
// address list is empty, it returns a zero address element instead.
func (a Addrs) First() Addr {
	if len(a) == 0 {
		return Addr{}
	}
	return a[0]
}
