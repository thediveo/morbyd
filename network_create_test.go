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
	"errors"
	"io"
	"time"

	types "github.com/docker/docker/api/types"
	"github.com/thediveo/morbyd/ipam"
	"github.com/thediveo/morbyd/net"
	"github.com/thediveo/morbyd/net/bridge"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/success"
)

var _ = Describe("creating custom networks", Ordered, func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Second).
			ShouldNot(HaveLeaked(goodgos))
	})

	It("rejects invalid options", func(ctx context.Context) {
		sess := Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=network.create.invopt")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		Expect(sess.CreateNetwork(ctx, "foobar",
			net.WithLabel(""))).Error().To(HaveOccurred())
	})

	It("creates a custom bridge network", func(ctx context.Context) {
		const name = "morbyd-custom-bridge-network"

		sess := Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=network.create.customnet")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		nw := Successful(sess.CreateNetwork(ctx, name,
			net.WithInternal(),
			bridge.WithBridgeName("brrr"),
			bridge.WithInterfacePrefix("brrr"),
			net.WithIPAM(ipam.WithPool(
				// https://en.wikipedia.org/wiki/Reserved_IP_addresses
				"0.0.1.0/24", // a.k.a. "local" or "this"
				ipam.WithRange("0.0.1.16/28")))))
		DeferCleanup(func(ctx context.Context) { _ = nw.Remove(ctx) })

		var buff safe.Buffer
		Expect(sess.Run(ctx, "busybox", run.WithCommand(
			"/bin/ip", "addr", "show"),
			run.WithCombinedOutput(io.MultiWriter(
				timestamper.New(GinkgoWriter), &buff)),
			run.WithAutoRemove(),
			run.WithNetwork(nw.Name))).Error().NotTo(HaveOccurred())
		// we should have been allocated an IPv4 from the "local" pool, and our
		// nif name should have the prefix we configured...
		Eventually(buff.String).Within(5 * time.Second).ProbeEvery(100 * time.Millisecond).
			Should(MatchRegexp(`(?s)\d+: brrr\d+@if\d+: .*inet 0\.0\.1\.(1[6-9]|2[0-9]|3[01])`))
	})

	It("returns an error when creation fails", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "NetworkCreate")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rec.NetworkCreate(Any, Any, Any).Return(types.NetworkCreateResponse{}, errors.New("error IJK305I"))

		Expect(sess.CreateNetwork(ctx, "foobar-telekomisch")).Error().To(MatchError(ContainSubstring("cannot create new network")))

	})

	It("returns an error when inspection after creation fails", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "NetworkCreate", "NetworkInspect", "NetworkRemove")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rec.NetworkCreate(Any, Any, Any).Return(types.NetworkCreateResponse{
			ID: "deadbeef",
		}, nil)
		rec.NetworkInspect(Any, Any, Any).Return(types.NetworkResource{}, errors.New("error IJK305I"))
		rec.NetworkRemove(Any, mock.Eq("deadbeef")).Return(nil)

		Expect(sess.CreateNetwork(ctx, "foobar-telekomisch")).Error().To(MatchError(ContainSubstring("cannot inspect newly created network")))
	})

})
