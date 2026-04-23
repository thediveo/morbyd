// Copyright 2026 Harald Albrecht.
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

package ensure

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ensuring things", func() {

	Context("maps", func() {

		It("creates only once", func() {
			var m map[string]string
			Expect(m).To(BeNil())

			Map(&m)
			Expect(m).NotTo(BeNil())
			m["foo"] = "bar"

			Map(&m)
			Expect(m).To(HaveLen(1))
			Expect(m).To(HaveKeyWithValue("foo", "bar"))
		})

	})

	Context("values", func() {

		It("creates only once", func() {
			var s *struct{ Canary bool }

			Value(&s)
			Expect(s.Canary).To(BeFalse())
			s.Canary = true
			s0 := s

			Value(&s)
			Expect(s.Canary).To(BeTrue())
			Expect(s).To(BeIdenticalTo(s0))
		})

	})

})
