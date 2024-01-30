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

package build

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gi "github.com/onsi/gomega/gstruct"
)

var _ = Describe("build args", func() {

	It("builds build args from a list of arg strings", func() {
		Expect(MakeBuildArgs("foo=bar", "foobar=", "baz")).To(And(
			HaveLen(3),
			HaveKeyWithValue("foo", gi.PointTo(Equal("bar"))),
			HaveKeyWithValue("foobar", gi.PointTo(Equal(""))),
			HaveKeyWithValue("baz", BeNil()),
		))
	})

	It("rejects invalid build args", func() {
		Expect(MakeBuildArgs("")).Error().To(HaveOccurred())
		Expect(MakeBuildArgs("=")).Error().To(HaveOccurred())
	})

})
