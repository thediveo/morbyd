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

package identity

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("user and group identities", func() {

	DescribeTable("user (with group) principals, not principles",
		func(p any, expected string) {
			var actual string
			switch v := p.(type) {
			case string:
				actual = WithUser(v)
			case int:
				actual = WithUser(v)
			}
			Expect(actual).To(Equal(expected))
		},
		Entry("clear", "", ""),
		Entry("user name", "maroding_marble", "maroding_marble"),
		Entry("user name", 1000, "1000"),
		Entry("user:group names", "peter:paul-marry", "peter:paul-marry"),
	)

	DescribeTable("user/group principals",
		func(pu string, pg any, expected string) {
			var actual string
			actual = WithUser(pu)
			switch v := pg.(type) {
			case string:
				actual = WithGroup(actual, v)
			case int:
				actual = WithGroup(actual, v)
			}
			Expect(actual).To(Equal(expected))
		},
		Entry("clear group", "1000:6666", "", "1000"),
		Entry("replace group", "1000:6666", "blue_breach", "1000:blue_breach"),
		Entry("replace group (int)", "1000:6666", 42, "1000:42"),
		Entry("add group", "1000", "blue_breach", "1000:blue_breach"),
	)
})
