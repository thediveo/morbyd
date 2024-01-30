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
	"bytes"
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	"github.com/thediveo/morbyd/session"
	. "github.com/thediveo/success"
)

var _ = Describe("pulling images", Ordered, func() {

	var sess *Session

	BeforeAll(func(ctx context.Context) {
		sess = Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=")))
		DeferCleanup(func(ctx context.Context) {
			// not strictly necessary as we're doing it anyway after each
			// individual test in order to check for leaked go routines.
			sess.Close(ctx)
		})
	})

	BeforeEach(func(ctx context.Context) {
		goodgos := Goroutines()
		DeferCleanup(func() {
			sess.Close(ctx)
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
		})
	})

	It("pulls an image", func(ctx context.Context) {
		var buff bytes.Buffer
		Expect(sess.PullImage(ctx, "busybox:latest", WithPullImageOutput(&buff))).To(Succeed())
		Expect(buff.String()).To(ContainSubstring("latest: Pulling from library/busybox"))
	})

	It("reports errors", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		Expect(sess.PullImage(ctx, "")).NotTo(Succeed())
	})

})
