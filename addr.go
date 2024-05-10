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
	"strconv"
)

// Addr represents a non-zero IP address together with a transport-layer port
// and the particular transport-layer protocol.
type Addr struct {
	l4proto string
	ip      net.IP
	port    uint16
}

// Addrs is a list of Addr elements, providing helper functions on top.
type Addrs []Addr

var _ (net.Addr) = (*Addr)(nil)

// NewAddr returns an Addr element.
func NewAddr(ip net.IP, port uint16, l4proto string) Addr {
	return Addr{
		l4proto: l4proto,
		ip:      ip,
		port:    port,
	}
}

// Network returns the “name” of the “network” the service is on, either “tcp”
// or “udp”.
func (a Addr) Network() string { return a.l4proto }

// String returns the address with port, such as “127.0.0.1:123” or “[::1]:123”.
// It returns "" for a zero Addr.
func (a Addr) String() string {
	if a.l4proto == "" {
		return ""
	}
	if ip := a.ip.To4(); ip != nil {
		return ip.String() + ":" + strconv.FormatUint(uint64(a.port), 10)
	}
	return "[" + a.ip.String() + "]:" + strconv.FormatUint(uint64(a.port), 10)
}

var ipv4loopback = net.ParseIP("127.0.0.1").To4()

// UnspecifiedAsLoopback returns the IPv4 or IPv6 loopback address if this Addr
// is unspecified, otherwise Addr unmodified. Use UnspecifiedAsLoopback to
// ensure to always get a non-unspecified destination address of a published
// container port.
func (a Addr) UnspecifiedAsLoopback() Addr {
	if a.l4proto == "" {
		return Addr{}
	}
	if !a.ip.IsUnspecified() {
		return a
	}
	if a.ip.To4() != nil {
		return Addr{
			l4proto: a.l4proto,
			ip:      ipv4loopback,
			port:    a.port,
		}
	}
	return Addr{
		l4proto: a.l4proto,
		ip:      net.IPv6loopback,
		port:    a.port,
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
