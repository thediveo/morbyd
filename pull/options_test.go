// Copyright 2025 Harald Albrecht.
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

package pull

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("pull image options", func() {

	It("processes pull image options", func() {
		popts := Options{}
		for _, opt := range []Opt{
			WithAllTags(),
			WithPlatform("sunkenos/sparq"),
			WithRegistryAuth("deadfoobar"),
			WithOutput(io.Discard),
		} {
			Expect(opt(&popts)).NotTo(HaveOccurred())
		}
		Expect(popts.All).To(BeTrue())
		Expect(popts.Platform).NotTo(BeNil())
		Expect(popts.Platform).To(Equal("sunkenos/sparq"))
		Expect(popts.Out).To(BeIdenticalTo(io.Discard))
	})

	It("rejects invalid platforms", func() {
		popts := Options{}
		Expect(WithPlatform("z80")(&popts)).NotTo(Succeed())
	})

})
