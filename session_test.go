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
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

var _ = Describe("test sessions", func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
		})
	})

	It("reports an error when the client cannot be created", func(ctx context.Context) {
		origdockerhost := os.Getenv(client.EnvOverrideHost)
		defer os.Setenv(client.EnvOverrideHost, origdockerhost)
		// There really aren't that many ways to trigger an error when creating
		// a Docker host, but overriding the host with an invalid protocol URL
		// is a rare option...
		os.Setenv(client.EnvOverrideHost, "lalala://")
		Expect(NewSession(ctx)).Error().To(HaveOccurred())
	})

	It("reports an option error", func(ctx context.Context) {
		Expect(NewSession(ctx, session.WithLabel("="))).Error().To(MatchError(
			MatchRegexp(`cannot create new test session.*label must be in format`)))
	})

	When("auto-cleaning", func() {

		It("skips auto-cleaning without a label set", Serial, func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl)))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})

			var buff safe.Buffer
			cntr := Successful(sess.Run(ctx, "busybox",
				run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; echo \"OK\"; while true; do sleep 1; done"),
				run.WithAutoRemove(),
				run.WithCombinedOutput(io.MultiWriter(&buff, timestamper.New(GinkgoWriter)))))
			DeferCleanup(func(ctx context.Context) {
				cntr.Kill(ctx)
			})
			Eventually(buff.String).Within(5 * time.Second).ProbeEvery(100 * time.Millisecond).
				Should(ContainSubstring("OK"))
			sess.AutoClean(ctx)
			Consistently(cntr.Refresh).WithContext(ctx).Within(3 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaventFoundContainer())
		})

		It("silently handles API network list errors", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "NetworkList")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.NetworkList(Any, Any).
				Return(nil, errors.New("error IJK305I")) // ...real programmers ;)
			sess.autoClean(ctx, "test.foo=bar")
		})

		It("silently handles API network removal errors", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "NetworkList", "NetworkRemove")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.NetworkList(Any, Any).
				Return([]types.NetworkResource{
					{ID: "42"},
					{ID: "666"},
				}, nil)
			rec.NetworkRemove(Any, mock.Eq("42")).
				Times(1).
				Return(nil)
			rec.NetworkRemove(Any, mock.Eq("666")).
				Times(1).
				Return(errors.New("error IJK305I"))
			sess.autoClean(ctx, "test.foo=bar")
		})

	})

	Context("with a session", Ordered, func() {

		var sess *Session

		BeforeEach(func(ctx context.Context) {
			sess = Successful(NewSession(ctx,
				session.WithAutoCleaning("test.morbyd=session")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
		})

		When("looking up containers", func() {

			It("reports an error for a non-existing ID", func(ctx context.Context) {
				Expect(sess.Container(ctx, "no-one-should-create-containers-with-this-name-sherly")).Error().
					To(HaventFoundContainer())
			})

			It("returns a new *Container", func(ctx context.Context) {
				name := "morbyd-test-session-container"
				cntr := Successful(sess.Run(ctx, "busybox",
					run.WithName(name),
					run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; while true; do sleep 1; done"),
					run.WithAutoRemove(),
					run.WithCombinedOutput(timestamper.New(GinkgoWriter))))
				c := Successful(sess.Container(ctx, name))
				Expect(c.Name).To(Equal(name))
				Expect(c.ID).To(Equal(cntr.ID))
				Expect(c.Session).NotTo(BeNil())
				Expect(c.Details).NotTo(BeZero())
			})

		})

		When("looking up networks", func() {

			It("reports an error for a non-existing ID", func(ctx context.Context) {
				Expect(sess.Network(ctx, "no-one-should-create-networks-with-this-name-sherly")).Error().
					To(HaventFoundNetwork())
			})

			It("returns a new *Network for an existing network", func(ctx context.Context) {
				name := "morbyd-test-session-network"
				netw := Successful(sess.CreateNetwork(ctx, name))
				n := Successful(sess.Network(ctx, name))
				Expect(n.Name).To(Equal(name))
				Expect(n.ID).To(Equal(netw.ID))
				Expect(n.Session).NotTo(BeNil())
				Expect(n.Details).NotTo(BeZero())
			})

		})

	})

	It("detects Docker Desktop platform", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ServerVersion")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rec.ServerVersion(Any).Return(types.Version{
			Platform: struct{ Name string }{"foobar Desktop"},
		}, nil)

		Expect(sess.IsDockerDesktop(ctx)).To(BeTrue())
	})

})
