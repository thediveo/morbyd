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
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/thediveo/morbyd/net"
	"golang.org/x/exp/maps"
)

// CreateNetwork creates a new “custom” Docker network using the specified
// configuration options.
//
// Notable configuration options:
//   - [net.WithInternal] marks the new network as “[internal]”.
//   - [net.WithDriver] allows setting a different Docker network driver, such
//     as “macvlan” or “ipvlan”.
//   - see [github.com/thediveo/morbyd/net/bridge],
//     [github.com/thediveo/morbyd/net/macvlan], and
//     [github.com/thediveo/morbyd/net/ipvlan] for further, drive-specific
//     configuration options.
//
// See also: [docker network create]
//
// [docker network create]: https://docs.docker.com/engine/reference/commandline/network_create/
func (s *Session) CreateNetwork(ctx context.Context, name string, opts ...net.Opt) (*Network, error) {
	nopts := net.Options{
		CheckDuplicate: true, // client now enforces this even before API v1.44
		Labels:         map[string]string{},
	}
	maps.Copy(nopts.Labels, s.opts.Labels)
	for _, opt := range opts {
		if err := opt((*net.Options)(&nopts)); err != nil {
			return nil, err
		}
	}

	createResp, err := s.moby.NetworkCreate(ctx, name, types.NetworkCreate(nopts))
	if err != nil {
		return nil, fmt.Errorf("cannot create new network %q, reason: %w",
			name, err)
	}

	detailsResp, err := s.moby.NetworkInspect(ctx, createResp.ID, types.NetworkInspectOptions{
		Verbose: true,
	})
	if err != nil {
		_ = s.moby.NetworkRemove(ctx, createResp.ID)
		return nil, fmt.Errorf("cannot inspect newly created network %q, reason: %w",
			name, err)
	}

	n := Network{
		Name:    name,
		ID:      createResp.ID,
		Session: s,
		Details: detailsResp,
	}
	return &n, nil
}
