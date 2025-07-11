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
	"math/rand"
	"time"

	types "github.com/docker/docker/api/types"
	container "github.com/docker/docker/api/types/container"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("containers", Ordered, func() {

	var sess *Session

	BeforeAll(func(ctx context.Context) {
		sess = Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=container")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
	})

	It("waits for a container to finish", func(ctx context.Context) {
		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithCommand("/bin/sleep", "5s"),
			run.WithAutoRemove(),
			run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
		))
		start := time.Now()
		Expect(cntr.Wait(ctx)).To(Succeed())
		Expect(time.Since(start)).To(BeNumerically(">=", 4*time.Second))
	})

	It("pauses and unpauses a container", func(ctx context.Context) {
		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithAutoRemove(),
			run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
		))
		Expect(cntr.PID(ctx)).Error().NotTo(HaveOccurred())

		Expect(cntr.Pause(ctx)).To(Succeed())
		Expect(cntr.Refresh(ctx)).To(Succeed())
		Expect(cntr.Details.State.Paused).To(BeTrue())

		Expect(cntr.Unpause(ctx)).To(Succeed())
		Expect(cntr.Refresh(ctx)).To(Succeed())
		Expect(cntr.Details.State.Paused).To(BeFalse())
	})

	It("stops a container cooperatively", func(ctx context.Context) {
		var buff safe.Buffer

		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; echo \"OK\"; while true; do sleep 1; done"),
			run.WithAutoRemove(),
			run.WithCombinedOutput(io.MultiWriter(&buff, timestamper.New(GinkgoWriter))),
		))
		Eventually(buff.String).Within(5 * time.Second).ProbeEvery(100 * time.Millisecond).
			Should(ContainSubstring("OK"))
		cntr.Stop(ctx)
		Eventually(cntr.Refresh).WithContext(ctx).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
			Should(HaventFoundContainer())
	})

	It("kills a container without mercy", func(ctx context.Context) {
		var buff safe.Buffer

		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; echo \"OK\"; while true; do sleep 1; done"),
			run.WithCombinedOutput(io.MultiWriter(&buff, timestamper.New(GinkgoWriter))),
		))
		Eventually(buff.String).Within(5 * time.Second).ProbeEvery(100 * time.Millisecond).
			Should(ContainSubstring("OK"))
		cntr.Kill(ctx)
		Eventually(cntr.Refresh).WithContext(ctx).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
			Should(HaventFoundContainer())
	})

	It("returns an abbreviated container ID", func() {
		c := &Container{}
		Expect(c.AbbreviatedID()).To(Equal(""))

		hexdigits := "0123456789ABCDEF"
		id := make([]byte, 64)
		for idx := range id {
			id[idx] = hexdigits[rand.Intn(len(hexdigits))]
		}

		c = &Container{ID: string(id)}
		Expect(c.AbbreviatedID()).To(Equal(string(id)[:AbbreviatedIDLength]))
	})

	It("returns an error when container refresh fails", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ContainerInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{}, errors.New("error IJK305I"))

		cntr := &Container{Session: sess, ID: "bad1dea"}
		Expect(cntr.Refresh(ctx)).Error().To(MatchError(ContainSubstring("cannot refresh details of container")))
	})

	It("returns an error when waiting fails", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ContainerWait")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		errch := make(chan error, 1)
		rec.ContainerWait(Any, Any, Any).Return(make(chan container.WaitResponse), errch)

		errch <- errors.New("error IJK305I")
		cntr := &Container{
			Session: sess,
			Name:    "foobar",
			ID:      "deadbeefc0011dea",
		}
		Expect(cntr.Wait(ctx)).Error().To(MatchError(ContainSubstring("waiting for container")))

	})

	It("renames a container", func(ctx context.Context) {
		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithAutoRemove(),
			run.WithName("test_foo"),
			run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; while true; do sleep 1; done")))
		Expect(cntr.PID(ctx)).Error().NotTo(HaveOccurred())
		Expect(cntr.Rename(ctx, "test_bar")).To(Succeed())
		Expect(cntr.Details.Name).To(Equal("/test_bar"))
		cntr.Kill(ctx)
	})

})
