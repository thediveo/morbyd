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
	"time"

	"github.com/docker/docker/api/types"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/success"
)

var _ = Describe("getting container PIDs", Ordered, func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Second).
			ShouldNot(HaveLeaked(goodgos))
	})

	Context("mocked", func() {

		It("retries until PID becomes available", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{},
			}, nil)
			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Pid: 42,
					},
				},
			}, nil)

			cntr := &Container{Session: sess, ID: "bad1dea"}
			Expect(cntr.PID(ctx)).To(Equal(42))
		})

		It("waits for a restart", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Dead:       true,
						Restarting: true,
					},
				},
			}, nil)
			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Pid: 42,
					},
				},
			}, nil)

			cntr := &Container{Session: sess, ID: "bad1dea"}
			Expect(cntr.PID(ctx)).To(Equal(42))
		})

		It("gives up when there's no chance of a restart", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{},
			}, nil)
			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						OOMKilled: true,
					},
				},
			}, nil)

			cntr := &Container{Session: sess, ID: "bad1dea"}
			Expect(cntr.PID(ctx)).Error().To(HaveOccurred())
		})

		It("reports an error when container cannot be inspected", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{}, errors.New("error IJK305I"))

			cntr := &Container{Session: sess, ID: "bad1dea"}
			Expect(cntr.PID(ctx)).Error().To(HaveOccurred())
		})

		It("returns when its nap gets cancelled", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			// Ignore the cancelled context so we can get to the short nap attack.
			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{},
			}, nil)
			ctx, cancel := context.WithCancel(ctx)
			cancel()

			cntr := &Container{Session: sess, ID: "bad1dea"}
			Expect(cntr.PID(ctx)).Error().To(HaveOccurred())
		})

	})

	Context("handling a failed container", func() {

		It("doesn't wait endless for PID of failed container", func(ctx context.Context) {
			sess := Successful(NewSession(ctx, session.WithAutoCleaning("test.morbid=pid")))
			DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

			By("creating a crashed container")
			cntr := Successful(sess.Run(ctx,
				"busybox",
				run.WithCommand("/bin/sh", "-c", "this feels wrong"),
				run.WithAutoRemove(),
				run.WithCombinedOutput(GinkgoWriter)))
			By("(not) getting PID of crashed container")
			Eventually(func(ctx context.Context) error {
				_, err := cntr.PID(ctx)
				return err
			}).WithContext(ctx).
				Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(Succeed())
		})

	})

})
