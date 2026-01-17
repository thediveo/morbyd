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
	"io"

	"github.com/docker/docker/api/types/build"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gs "github.com/onsi/gomega/gstruct"
)

var _ = Describe("image build options", func() {

	It("processes image build options", func() {
		bios := Options{}
		for _, opt := range []Opt{
			WithTag("fool"),
			WithTag("fool:oncemore"),

			WithBuildArg("foobar="),
			WithBuildArgs("foo=bar", "baz"),

			WithDockerfile("Duckerphile"),
			WithLabel("foo=bar"),
			WithLabels("fool=bar", "jekyll=hyde"),
			WithoutCache(),
			WithPullAlways(),
			WithRemoveIntermediateContainers(),
			WithAlwaysRemoveIntermediateContainers(),
			WithSquash(),

			WithOutput(io.Discard),
		} {
			Expect(opt(&bios)).NotTo(HaveOccurred())
		}
		Expect(bios.Tags).To(ConsistOf("fool", "fool:oncemore"))
		Expect(bios.BuildArgs).To(And(
			HaveLen(3),
			HaveKeyWithValue("foo", gs.PointTo(Equal("bar"))),
			HaveKeyWithValue("foobar", gs.PointTo(Equal(""))),
			HaveKeyWithValue("baz", BeNil()),
		))
		Expect(bios.Dockerfile).To(Equal("Duckerphile"))
		Expect(bios.Labels).To(And(
			HaveLen(3),
			HaveKeyWithValue("foo", "bar"),
			HaveKeyWithValue("fool", "bar"),
			HaveKeyWithValue("jekyll", "hyde"),
		))
		Expect(bios.NoCache).To(BeTrue())
		Expect(bios.PullParent).To(BeTrue())
		Expect(bios.Remove).To(BeTrue())
		Expect(bios.ForceRemove).To(BeTrue())
		Expect(bios.Squash).To(BeTrue())
		Expect(bios.Out).To(BeIdenticalTo(io.Discard))

		Expect(WithOpts(build.ImageBuildOptions{})(&bios)).NotTo(HaveOccurred())
		Expect(bios.Dockerfile).To(BeEmpty())
		Expect(bios.Out).To(BeIdenticalTo(io.Discard))
	})

	It("sets up the build args map and rejects invalid build args", func() {
		var opts Options
		Expect(WithBuildArgs("foo", "")(&opts)).Error().To(HaveOccurred())
	})

	It("sets up the labels map and rejects invalid labels", func() {
		var opts Options
		Expect(WithLabels("foo=", "")(&opts)).Error().To(HaveOccurred())
	})

})
