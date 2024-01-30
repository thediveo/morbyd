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

import "github.com/docker/docker/api/types/network"

// PoolOpt is a configuration option for an IP address pool.
type PoolOpt func(*Pool) error

// Pool represents an IP address pool, consisting of an IP subnet, and
// optionally an IP range inside the subnet to hand out addresses from, an
// optional default router IP address, and optional blocked (auxiliary) IP
// addresses.
type Pool network.IPAMConfig

// makePool returns a new IPAM Pool object, allocating IP addresses from the
// specified subnet, or optionally from a sub-range within the subnet. Other
// options allow further customization.
func makePool(subnet string, opts ...PoolOpt) (Pool, error) {
	p := Pool{
		Subnet: subnet,
	}
	for _, opt := range opts {
		if err := opt(&p); err != nil {
			return Pool{}, err
		}
	}
	return p, nil
}

// WithRange specifies a (sub) range of IP addresses within the Pool's subnet to
// allocate the addresses from (instead of allocating from the whole subnet
// range by default).
func WithRange(iprange string) PoolOpt {
	return func(p *Pool) error {
		p.IPRange = iprange
		return nil
	}
}

// WithGateway specifies the default router's (“gateway” in IPv4 parlance) IP
// address.
func WithGateway(gw string) PoolOpt {
	return func(p *Pool) error {
		p.Gateway = gw
		return nil
	}
}

// WithAuxAddress informs the IPAM driver of an IP address already in use in the
// network, so the driver must not allocate it to a container. In addition, this
// auxiliary IP address is assigned the specified hostname.
func WithAuxAddress(hostname string, ip string) PoolOpt {
	return func(p *Pool) error {
		if p.AuxAddress == nil {
			p.AuxAddress = map[string]string{}
		}
		p.AuxAddress[hostname] = ip
		return nil
	}
}
