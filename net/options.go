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

package net

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/thediveo/morbyd/internal/ipamint"
	"github.com/thediveo/morbyd/internal/netint"
	"github.com/thediveo/morbyd/ipam"
	lbls "github.com/thediveo/morbyd/labels"
)

// Opt is a configuration option when creating a custom Docker network using
// [github.com/thediveo/morbyd.Session.NewNetwork].
type Opt func(*Options) error

// Options represents the configuration options when creating a custom Docker
// network, including [ipam] configuration options.
type Options types.NetworkCreate

// WithDriver specifies the network driver (plugin) to use when creating a new
// Docker network. If left unspecified, it automatically defaults to Docker's
// “bridge” driver.
func WithDriver(name string) Opt {
	return func(o *Options) error {
		o.Driver = name
		return nil
	}
}

// WithIPAM specifies the particular IPAM driver configuration to use for
// allocating IP addresses to containers getting attached to this network. See
// also [ipam.Driver].
func WithIPAM(opts ...ipam.IPAMOpt) Opt {
	return func(o *Options) error {
		drv, err := ipamint.MakeIPAM(opts...)
		if err != nil {
			return err
		}
		o.IPAM = (*network.IPAM)(&drv)
		return nil
	}
}

// WithInternal sets the Docker network to be created as “internal”.
func WithInternal() Opt {
	return func(o *Options) error {
		o.Internal = true
		return nil
	}
}

// WithIPv6 enables IPv6 for the custom Docker network \o/.
func WithIPv6() Opt {
	return func(o *Options) error {
		o.EnableIPv6 = true
		return nil
	}
}

// WithLabel adds a label in “KEY=VALUE” to the custom Docker network.
func WithLabel(label string) Opt {
	return func(o *Options) error {
		netint.EnsureLabelsMap((*types.NetworkCreate)(o))
		return lbls.Labels(o.Labels).Add(label)
	}
}

// WithLabels adds multiple key-value labels to Docker network.
func WithLabels(labels ...string) Opt {
	return func(o *Options) error {
		netint.EnsureLabelsMap((*types.NetworkCreate)(o))
		for _, label := range labels {
			if err := lbls.Labels(o.Labels).Add(label); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithOption adds a driver option in “KEY=VALUE” format.
func WithOption(opt string) Opt {
	return func(o *Options) error {
		netint.EnsureOptionsMap((*types.NetworkCreate)(o))
		return lbls.Labels(o.Options).Add(opt)
	}
}
