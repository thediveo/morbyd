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

package labels

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("label maps", func() {

	It("builds label maps only from valid labels, and reject invalid labels", func() {
		Expect(MakeLabels("foo=bar", "foobar=")).To(And(
			HaveLen(2),
			HaveKeyWithValue("foo", "bar"),
			HaveKeyWithValue("foobar", ""),
		))
		Expect(MakeLabels("")).Error().To(HaveOccurred())
		Expect(MakeLabels("=")).Error().To(HaveOccurred())
		Expect(MakeLabels("=bar")).Error().To(HaveOccurred())
	})

})
