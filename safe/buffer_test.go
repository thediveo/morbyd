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

package safe

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("concurrency-safe buffer", func() {

	// This test must be run with the race detector enabled, as otherwise we
	// cannot ensure that any hiccup in our minimalist implementation is
	// detected.
	It("is safe for racing", func() {
		if !raceEnabled {
			Fail("-race required")
		}

		const message = "D'OH!"

		var buff Buffer
		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(buff.Write([]byte(message))).Error().To(Succeed())
		}()
		Eventually(done).Within(time.Second).ProbeEvery(10 * time.Millisecond).
			Should(BeClosed())
		Expect(buff.String()).To(Equal(message))
	})

})
