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

	"github.com/docker/docker/api/types/container"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/success"
)

var _ = Describe("waiting for commands executing inside containers", Ordered, func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Second).
			ShouldNot(HaveLeaked(goodgos))
	})

	It("reports cancelled contexts", func(ctx context.Context) {
		sess := Successful(NewSession(ctx))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ex := &ExecSession{
			Container: &Container{
				Session: sess,
			},
		}
		Expect(ex.Wait(ctx)).Error().To(HaveOccurred())
	})

	It("reports when the introspection fails", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ContainerExecInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rec.ContainerExecInspect(Any, Any).Return(container.ExecInspect{}, errors.New("error IJK305I"))

		ex := &ExecSession{
			Container: &Container{
				Session: sess,
			},
			done: make(chan struct{}),
		}
		close(ex.done)
		Expect(ex.Wait(ctx)).Error().To(HaveOccurred())
	})

	It("reports when the command is alive despite being dead", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ContainerExecInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rec.ContainerExecInspect(Any, Any).Return(container.ExecInspect{
			Running: true,
		}, nil)

		ex := &ExecSession{
			Container: &Container{
				Session: sess,
			},
			done: make(chan struct{}),
		}
		close(ex.done)
		Expect(ex.Wait(ctx)).Error().To(HaveOccurred())
	})

})
