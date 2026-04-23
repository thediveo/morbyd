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

package remove

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func opts(opts ...Opt) Options {
	GinkgoHelper()
	o := Options{}
	for _, opt := range opts {
		Expect(opt(&o)).To(Succeed())
	}
	return o
}

var _ = Describe("remove image options", func() {

	It("processes remove image options", func() {
		o := opts(
			WithSchwartz(),
			WithPruneChildren(),
			WithPlatform("linux/arm64"),
			WithPlatform("linux/amd64"),
		)
		Expect(o.Force).To(BeTrue())
		Expect(o.PruneChildren).To(BeTrue())
		Expect(o.Platforms).To(ConsistOf(
			ocispec.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			ocispec.Platform{
				OS:           "linux",
				Architecture: "arm64",
			},
		))
	})

})
