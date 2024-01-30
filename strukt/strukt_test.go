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

package strukt

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("unmarshalling delemited-field strings into structs", func() {

	It("rejects non-structs", func() {
		var answer int
		Expect(Unmarshal("42", ":", answer)).To(MatchError(
			ContainSubstring("expected *struct { ... }, got int")))
		Expect(Unmarshal("42", ":", &answer)).To(MatchError(
			ContainSubstring("expected *struct { ... }, got *int")))
	})

	It("expects a *struct", func() {
		var s struct {
			answer string
		}
		Expect(Unmarshal("42", ":", s)).To(MatchError(
			ContainSubstring("expected *struct { ... }, got struct {")))
	})

	It("rejects too many fields", func() {
		var s struct {
			F1 string
		}
		Expect(Unmarshal("42:666", ":", &s)).To(MatchError(
			ContainSubstring("too many fields for struct type *struct { F1 string }")))
	})

	It("rejects structs with unsettable fields", func() {
		var s struct {
			f string
		}
		Expect(Unmarshal("42", ":", &s)).To(MatchError(
			ContainSubstring("cannot set field f")))
	})

	It("rejects structs with non-string fields", func() {
		var s struct {
			F bool
		}
		Expect(Unmarshal("42", ":", &s)).To(MatchError(
			ContainSubstring("expected field F to have type string, got bool")))
	})

	It("unmarshals", func() {
		var s struct {
			F1 string
			F2 string
		}
		Expect(Unmarshal("42", ":", &s)).To(Succeed())
		Expect(s.F1).To(Equal("42"))
		Expect(s.F2).To(BeEmpty())

		Expect(Unmarshal("42:666", ":", &s)).To(Succeed())
		Expect(s.F2).To(Equal("666"))
	})

})
