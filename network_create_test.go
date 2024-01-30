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
	"io"
	"time"

	"github.com/thediveo/morbyd/ipam"
	"github.com/thediveo/morbyd/net"
	"github.com/thediveo/morbyd/net/bridge"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("create custom networks", Ordered, func() {

	var sess *Session

	BeforeAll(func(ctx context.Context) {
		sess = Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
	})

	It("rejects invalid options", func(ctx context.Context) {
		Expect(sess.CreateNetwork(ctx, "foobar",
			net.WithLabel(""))).Error().To(HaveOccurred())
	})

	It("creates a custom bridge network", func(ctx context.Context) {
		const name = "morbyd-custom-bridge-network"

		nw := Successful(sess.CreateNetwork(ctx, name,
			net.WithInternal(),
			bridge.WithBridgeName("brrr"),
			bridge.WithInterfacePrefix("brrr"),
			net.WithIPAM(ipam.WithPool(
				// https://en.wikipedia.org/wiki/Reserved_IP_addresses
				"0.0.1.0/24", // a.k.a. "local" or "this"
				ipam.WithRange("0.0.1.16/28")))))

		var buff safe.Buffer
		sess.Run(ctx, "busybox", run.WithCommand(
			"/bin/ip", "addr", "show"),
			run.WithCombinedOutput(io.MultiWriter(
				timestamper.New(GinkgoWriter), &buff)),
			run.WithAutoRemove(),
			run.WithNetwork(nw.Name))
		// we should have been allocated an IPv4 from the "local" pool, and our
		// nif name should have the prefix we configured...
		Eventually(buff.String).Within(5 * time.Second).ProbeEvery(100 * time.Millisecond).
			Should(MatchRegexp(`(?s)\d+: brrr\d+@if\d+: .*inet 0\.0\.1\.(1[6-9]|2[0-9]|3[01])`))
	})

})
