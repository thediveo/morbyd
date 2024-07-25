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
	context "context"
	io "io"
	"net/http"

	"github.com/thediveo/morbyd/net"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("published container ports", Ordered, func() {

	It("publishes container ports and talks to them", func(ctx context.Context) {
		sess := Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=container.port")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		By("spinning up an http serving busybox with published ports")
		cntr := Successful(sess.Run(ctx,
			"busybox",
			run.WithCommand("/bin/sh", "-c",
				`echo "DOH!" > index.html && httpd -v -f -p 1234`),
			run.WithAutoRemove(),
			run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
			// It's not possible (anymore) to publish the port both at the
			// unspecified IP/IPv6 addresses, as well as on the IPv6 loopback
			// address: this will fail with a bind error in the "userland
			// proxy". Thus, we publish on a fixed port on IPv6 loopback to work
			// around this restriction.
			run.WithPublishedPort("1234"),
			run.WithPublishedPort("[::1]:1234:1234/tcp"),
		))

		svcAddrs := cntr.PublishedPort("1234")
		Expect(svcAddrs).To(ConsistOf(
			And(
				HaveField("Network()", "tcp"),
				MatchRegexp(`0\.0\.0\.0:\d+`),
			),
			And(
				HaveField("Network()", "tcp"),
				MatchRegexp(`\[::\]:\d+`),
			),
			And(
				HaveField("Network()", "tcp"),
				MatchRegexp(`\[::1\]:\d+`),
			),
		))

		By("asking service for the magic phrase")
		svcAddrPort := svcAddrs.Any().UnspecifiedAsLoopback().String()
		Expect(svcAddrPort).NotTo(BeEmpty())
		resp := Successful(http.DefaultClient.Do(
			Successful(http.NewRequest(
				http.MethodGet, "http://"+svcAddrPort+"/", nil)).WithContext(ctx)))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()
		Expect(string(Successful(io.ReadAll(resp.Body)))).To(Equal("DOH!\n"))
	})

	It("publishes ports on an IPv6 custom Docker network", func(ctx context.Context) {
		sess := Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=container.portv6")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		v6net := Successful(sess.CreateNetwork(ctx, "morbyd-v6notwork",
			net.WithIPv6()))

		By("spinning up an http serving busybox with published ports")
		cntr := Successful(sess.Run(ctx,
			"busybox",
			run.WithCommand("/bin/sh", "-c",
				`echo "DOH!" > index.html && httpd -v -f -p 1235`),
			run.WithAutoRemove(),
			run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
			run.WithNetwork(v6net.ID),
			run.WithPublishedPort("[::1]:1235/tcp"),
		))

		svcAddrs := cntr.PublishedPort("1235")
		Expect(svcAddrs).To(ConsistOf(
			And(
				HaveField("Network()", "tcp"),
				MatchRegexp(`\[::1\]:\d+`),
			),
		))
	})

})
