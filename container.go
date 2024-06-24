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
	"net"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

// AbbreviatedIDLength defines the number of hex digits of a container ID to
// show in error and log messages.
const AbbreviatedIDLength = 10

// Container represents a Docker container, providing notable operations
// specific to it:
//
//   - [Container.IP] returns an host-internal IP address where the container
//     can be reached.
//   - [Container.Exec] to execute a command inside the container.
//   - [Container.PID] to retrieve the PID of the container's initial process.
//   - [Container.Stop] to stop the container by sending it the configured
//     signal (defaults to SIGTERM).
//   - [Container.Kill] to forcefully kill the container using SIGKILL.
type Container struct {
	Name    string
	ID      string
	Session *Session
	Details types.ContainerJSON // inspection information after start.
}

// Refresh the details about this container, or return an error in case
// refreshing fails.
func (c *Container) Refresh(ctx context.Context) error {
	details, err := c.Session.moby.ContainerInspect(ctx, c.ID)
	if err != nil {
		return fmt.Errorf("cannot refresh details of container %q/%s, reason: %w",
			c.Name, c.AbbreviatedID(), err)
	}
	c.Details = details
	return nil
}

// IP returns an IP address (net.IP) of this container that can be used to reach
// the container from the host. If no suitable IP address can be found, IP
// return nil. IP ignores addresses on a MACVLAN network, as IP addresses on a
// MACVLAN network cannot reached from the host.
//
// NOTE: the container's IP address is usable without the need to (publicly)
// expose container ports on the host – which often is less than desirable in
// tests. However, with Docker Desktop the container IPs aren't directly
// reachable anymore as on plain Docker hosts, so in these cases you'll need to
// expose a container's exposable ports on (preferably) loopback.
func (c *Container) IP(ctx context.Context) net.IP {
	// The container's own list of networks it is attached to unfortunately
	// doesn't tell us what the driver is. However, we need to know in order to
	// correctly skip MACVLANs...
	for _, netw := range c.Details.NetworkSettings.Networks {
		details, err := c.Session.moby.NetworkInspect(ctx, netw.NetworkID, types.NetworkInspectOptions{
			Verbose: true,
		})
		if err != nil {
			continue
		}
		switch details.Driver {
		case "macvlan":
			continue
		case "host":
			// Note that a container with "net:host" cannot be connected to any
			// other network, so this is a sufficient response.
			return net.ParseIP("127.0.0.1")
		case "null": // a.k.a. "net:none"
			// Note that a container with "net:none" (lo only) cannot be
			// connected to any other network, so this is a sufficient response.
			return nil
		}
		if netw.IPAddress == "" {
			continue
		}
		return net.ParseIP(netw.IPAddress)
	}
	return nil
}

// PID of the initial container process, as seen by the container engine. In
// case the container is restarting, it waits for the next Doctor, erm,
// container incarnation to come online.
//
// Note to Docker Desktop users: the PID is only valid in the context of the
// Docker engine that in case of macOS runs in its own VM, and in case of WSL2
// in its own PID namespace in the same HyperV Linux VM.
func (c *Container) PID(ctx context.Context) (int, error) {
	for {
		inspRes, err := c.Session.moby.ContainerInspect(ctx, c.ID)
		if err != nil {
			return 0, fmt.Errorf("cannot determine PID of container %s/%q, reason: %w",
				c.Name, c.AbbreviatedID(), err)
		}
		// If we got a non-zero PID, then no worries and we just return that.
		if inspRes.State != nil && inspRes.State.Pid != 0 {
			return inspRes.State.Pid, nil
		}
		// We're either too early or too late, but we have to figure out which
		// one we're, because we don't want to hang around any further if there
		// is no chance of getting a PID in the near future...
		if inspRes.State != nil &&
			((inspRes.State.Dead || inspRes.State.OOMKilled) &&
				!inspRes.State.Restarting) {
			return 0, fmt.Errorf("cannot determine PID of container %s/%q in state %q",
				c.Name, c.AbbreviatedID(), inspRes.State.Status)
		}
		if err := Sleep(ctx, DefaultSleep); err != nil {
			return 0, fmt.Errorf("cannot determine PID of container %s/%q, reason: %w",
				c.Name, c.AbbreviatedID(), err)
		}
	}
}

// Stop the container by sending it a termination signal. Default is SIGTERM,
// unless changed using [run.WithStopSignal].
func (c *Container) Stop(ctx context.Context) {
	_ = c.Session.moby.ContainerStop(ctx, c.ID, container.StopOptions{})
}

// Wait for the container to finish, that is, become “not-running” in Docker API
// parlance. See also: [Docker's Client.ContainerWait].
//
// [Docker's Client.ContainerWait]: https://pkg.go.dev/github.com/docker/docker/client#Client.ContainerWait
func (c *Container) Wait(ctx context.Context) error {
	// Nota bene: errch is buffered with size 1. The wait result channel is
	// unbuffered though. ContainerWait EITHER sends an error OR a result, never
	// both. And in consequence it never sends a nil error.
	waitch, errch := c.Session.moby.ContainerWait(ctx, c.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errch:
		return fmt.Errorf("waiting for container %q/%s to finish failed, reason: %w",
			c.Name, c.AbbreviatedID(), err)
	case <-waitch:
		return nil
	}
}

// Kill the container forcefully and also remove its volumes.
func (c *Container) Kill(ctx context.Context) {
	_ = c.Session.moby.ContainerRemove(ctx, c.ID, container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
}

// AbbreviatedID returns an abbreviated container ID for use in error reporting
// in order to not report unwieldy long IDs.
func (c *Container) AbbreviatedID() string {
	if len(c.ID) <= AbbreviatedIDLength {
		return c.ID
	}
	return c.ID[:AbbreviatedIDLength]
}

// PublishedPort returns the host IP address(es) and port(s) that forward to the
// transport-layer port and protocol of this container, such as “1234” or
// “1234/tcp”. If the transport-layer protocol is left unspecified, “tcp” is
// assumed by default.
//
// If there is no such port (with protocol) published, PublishedPort returns an
// empty [Addrs] list.
//
// In order to easily connect to a published container service, a suitable IP
// address string including port number can be determined as follows:
//
//	// for instance, returns "127.0.0.1:32890"
//	svcAddrPort := cntr.PublishedPort("1234").Any().UnspecifiedAsLoopback().String()
func (c *Container) PublishedPort(portproto string) Addrs {
	if !strings.Contains(portproto, "/") {
		portproto += "/tcp"
	}
	_, l4proto, _ := strings.Cut(portproto, "/")
	addrs := Addrs{}
	for _, boundport := range c.Details.NetworkSettings.Ports[nat.Port(portproto)] {
		ip := net.ParseIP(boundport.HostIP)
		if ip == nil {
			continue
		}
		if ip4 := ip.To4(); ip4 != nil {
			ip = ip4 // compact IPv4 addresses
		}
		port, err := strconv.ParseUint(boundport.HostPort, 10, 16)
		if err != nil {
			continue
		}
		addrs = append(addrs, NewAddr(ip, uint16(port), l4proto))
	}
	return addrs
}
