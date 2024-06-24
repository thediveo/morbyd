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

package morbyd_test

import (
	"context"
	"fmt"
	"time"

	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/ipam"
	"github.com/thediveo/morbyd/net"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
)

// Create a new “custom” Docker network, then run an example container attached
// to this (purely internal) network.
//
// We start by creating a session using [NewSession]. Because unit tests may
// crash and leave test containers and networks behind, we enable
// “auto-cleaning” using the [session.WithAutoCleaning] option, passing it a
// unique label. This label can be either a unique key (“KEY=”) or a unique
// key-value pair (“KEY=VALUE”); either form is allowed, depending on how you
// like to structure and label your test containers and networks. Auto-cleaning
// runs automatically directly after session creation (to remove any left-overs
// from a previous test run) and then again when calling [Session.Close].
//
// Please note that for this testable example we need a deterministic container
// IP address assignment. In your tests, you most probably just need a working
// IP address, but not a particular one, so you won't need [net.WithIPAM] in
// most circumstance.
//
// In our special case, we create a custom Docker network with an IP address
// management (IPAM) pool of “0.0.1.0/24” that is a small part of the so-called
// “this” network defined in [RFC5735 section 3] and [RFC8190].
//
// The default “bridge” driver will automatically allocate the first available
// pool IP address “0.0.1.1” to the Linux kernel bridge, so the first container
// IP address will be “0.0.1.2”. Please note that the IP address “0.0.1.0” is
// the “subnet address” of this subnet and usually is reserved, so Docker's
// standard IPAM driver never assigns it.
//
// When the example container attached to this custom network (using
// [run.WithNetwork]) starts, it executes the following shell command that grabs
// information about the specified network interface “eth0” and then cuts out
// only the IPv4 address:
//
//	ip a sh dev eth0 | awk '/inet / { print $2 }'
//
// [RFC5735 section 3]: https://datatracker.ietf.org/doc/html/rfc5735#section-3
// [RFC8190]: https://datatracker.ietf.org/doc/html/rfc8190
//
// [internal]: https://docs.docker.com/compose/compose-file/06-networks/#internal
func ExampleSession_CreateNetwork() {
	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)
	defer cancel()

	sess, err := morbyd.NewSession(ctx,
		session.WithAutoCleaning("test.morbyd=example.session.network"))
	if err != nil {
		panic(err)
	}
	defer sess.Close(ctx)

	_ = sess.Client().NetworkRemove(ctx, "my-notwork")
	netw, err := sess.CreateNetwork(ctx, "my-notwork",
		net.WithInternal(),
		net.WithIPAM(ipam.WithPool("0.0.42.0/24")))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = netw.Remove(ctx)
	}()

	var out safe.Buffer
	container, err := sess.Run(ctx,
		"busybox",
		run.WithCommand("/bin/sh", "-c", "ip a sh dev eth0 | awk '/inet / { print $2 }'"),
		run.WithNetwork(netw.ID),
		run.WithAutoRemove(),
		run.WithCombinedOutput(&out))
	if err != nil {
		panic(err)
	}

	_ = container.Wait(ctx)
	fmt.Print(out.String())
	// Output: 0.0.42.2/24
}
