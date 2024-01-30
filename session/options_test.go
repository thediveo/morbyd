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

package session

import (
	"github.com/docker/docker/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test session options", func() {

	It("processes session options", func() {
		sessos := Options{}
		for _, opt := range []Opt{
			WithAutoCleaning("test=morbyd-session"),
			WithLabel("foo=bar"),
			WithLabels("fool=bar", "jekyll=hyde"),
			WithDockerOpts(client.WithHost("unix:///doh/run/docker.sock")),
		} {
			Expect(opt(&sessos)).To(Succeed())
		}
		Expect(sessos.AutoCleaningLabel).To(Equal("test=morbyd-session"))
		Expect(sessos.Labels).To(And(
			HaveLen(4),
			HaveKeyWithValue("test", "morbyd-session"),
			HaveKeyWithValue("foo", "bar"),
			HaveKeyWithValue("fool", "bar"),
			HaveKeyWithValue("jekyll", "hyde"),
		))
		Expect(sessos.DockerClientOpts).To(HaveLen(1))
	})

	It("reports errors when rejecting invalid session options", func() {
		sessos := Options{}
		Expect(WithAutoCleaning("")(&sessos)).To(MatchError(MatchRegexp(
			`auto cleaning label must be in format .*, got ""`)))
		Expect(WithAutoCleaning("=")(&sessos)).To(MatchError(MatchRegexp(
			`auto cleaning label must be in format .*, got "="`)))
		Expect(WithLabels("foo=bar", "=")(&sessos)).To(MatchError(MatchRegexp(
			`label must be in format .*, got "="`)))
	})

})
