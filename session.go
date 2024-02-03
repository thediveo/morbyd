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
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/thediveo/morbyd/moby"
	"github.com/thediveo/morbyd/session"
)

// Session represents a Docker API client connection, together with additional
// configuration options that are inherited to newly created images, containers,
// and networks.
type Session struct {
	opts session.Options
	moby moby.Client
}

// NewSession creates a new Docker client and test session, returning a Session
// object on success, or an error otherwise.
//
// When [sess.WithAutoCleaning] has been specified, then NewSession will then
// forcefully remove all containers and then networks matching the specified
// auto-cleaning label. In this case, [Session.Close] will then run a
// post-session cleaning.
//
// Note: the Docker client is created using the options [client.FromEnv] and
// [client.WithAPIVersionNegotiation].
func NewSession(ctx context.Context, opts ...session.Opt) (*Session, error) {
	s := &Session{}
	for _, opt := range opts {
		if err := opt(&s.opts); err != nil {
			return nil, fmt.Errorf("cannot create new test session, reason: %w",
				err)
		}
	}
	clientOpts := append([]client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}, s.opts.DockerClientOpts...)
	moby, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	s.moby = moby
	if s.opts.Wrapper != nil {
		s.moby = s.opts.Wrapper(s.moby)
	}
	s.AutoClean(ctx)
	return s, nil
}

// Client returns the Docker client used in this test session.
func (s *Session) Client() moby.Client { return s.moby }

// Close removes left-over containers and networks if auto-cleaning has been
// enabled, and then closes idle HTTP connections to the Docker daemon.
func (s *Session) Close(ctx context.Context) {
	s.AutoClean(ctx)
	s.moby.Close()
}

// AutoClean forcefully removes all left-over containers and networks that are
// labelled with the auto-cleaning label specified when creating this session.
// If no auto-cleaning label was specified, AutoClean simply returns, doing
// nothing. (Well, it does something: it returns ... but that is now too meta).
func (s *Session) AutoClean(ctx context.Context) {
	if s.opts.AutoCleaningLabel == "" {
		return
	}

	// Assemble a filter based on the auto-cleaning label.
	aclabel := s.opts.AutoCleaningLabel
	key, value, _ := strings.Cut(aclabel, "=")
	if value == "" {
		aclabel = key // just the key, no trailing "=" in case of a zero value.
	}
	f := filters.NewArgs(filters.Arg("label", aclabel))

	// List all matching containers and then kill them.
	cntrs, err := s.moby.ContainerList(ctx, container.ListOptions{
		Filters: f,
	})
	if err != nil {
		return
	}
	for _, cntr := range cntrs {
		err := s.moby.ContainerRemove(ctx, cntr.ID, container.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		})
		if err != nil {
			return
		}
	}

	// List all matching networks (which by now should not have any test
	// containers attached to them anymore) and then remove them.
	nets, err := s.moby.NetworkList(ctx, types.NetworkListOptions{
		Filters: f,
	})
	if err != nil {
		return
	}
	for _, net := range nets {
		err := s.moby.NetworkRemove(ctx, net.ID)
		if err != nil {
			return
		}
	}
}

// Container returns a *Container object for the specified name or ID if it
// exists, otherwise it returns an error. Please note that multiple calls for
// the same name or ID will return different *Container objects, as there is no
// caching.
func (s *Session) Container(ctx context.Context, nameID string) (*Container, error) {
	details, err := s.moby.ContainerInspect(ctx, nameID)
	if err != nil {
		return nil, err
	}
	cntr := &Container{
		Name:    details.Name[1:],
		ID:      details.ID,
		Session: s,
		Details: details,
	}
	return cntr, nil
}

// Network returns a *Network object for the specified name or ID if it exists,
// otherwise it returns an error. Please note that multiple calls for the same
// name or ID will return different *Network objects, as there is no caching.
func (s *Session) Network(ctx context.Context, nameID string) (*Network, error) {
	details, err := s.moby.NetworkInspect(ctx, nameID, types.NetworkInspectOptions{
		Verbose: true,
	})
	if err != nil {
		return nil, err
	}
	netw := &Network{
		Name:    details.Name,
		ID:      details.ID,
		Session: s,
		Details: details,
	}
	return netw, nil
}
