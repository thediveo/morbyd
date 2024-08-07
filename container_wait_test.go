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
	"time"

	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/success"
)

var _ = Describe("waiting for a container to terminate", Ordered, func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Second).
			ShouldNot(HaveLeaked(goodgos))
	})

	It("doesn't wait endless for failed container", func(ctx context.Context) {
		sess := Successful(NewSession(ctx, session.WithAutoCleaning("test.morbid=container.wait")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		By("creating a crashed container")
		cntr := Successful(sess.Run(ctx,
			"busybox",
			run.WithCommand("/bin/sh", "-c", "this feels wrong"),
			run.WithAutoRemove(),
			run.WithCombinedOutput(GinkgoWriter)))
		By("waiting for crashed container")
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		Expect(cntr.Wait(ctx)).To(Or(
			Succeed(),
			MatchError(
				MatchRegexp(`waiting for container ".+"/[[:xdigit:]]+ to finish failed, .+ No such container`))))
	})

})
