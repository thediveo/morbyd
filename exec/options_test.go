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

package exec

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

func opts(opts ...Opt) Options {
	GinkgoHelper()
	o := Options{}
	for _, opt := range opts {
		Expect(opt(&o)).To(Succeed())
	}
	return o
}

var _ = Describe("exec command options", func() {

	It("processes input/output options", func() {
		var (
			stdin  strings.Reader
			stdout bytes.Buffer
			stderr bytes.Buffer
		)

		Expect(opts()).To(And(
			HaveField("In", BeNil()),
			HaveField("Out", BeNil()),
			HaveField("Err", BeNil()),
		))

		Expect(opts(WithCombinedOutput(&stdout))).To(And(
			HaveField("Conf.Tty", BeFalse()),
			HaveField("In", BeNil()),
			HaveField("Out", BeIdenticalTo(&stdout)),
			HaveField("Err", BeIdenticalTo(&stdout)),
		))

		Expect(opts(WithDemuxedOutput(&stdout, &stderr))).To(And(
			HaveField("Conf.Tty", BeFalse()),
			HaveField("In", BeNil()),
			HaveField("Out", BeIdenticalTo(&stdout)),
			HaveField("Err", BeIdenticalTo(&stderr)),
		))

		Expect(opts(WithInput(&stdin))).To(And(
			HaveField("Conf.Tty", BeFalse()),
			HaveField("In", BeIdenticalTo(&stdin)),
			HaveField("Out", BeNil()),
			HaveField("Err", BeNil()),
		))
	})

	It("processes exec options", func() {
		exopts := opts(
			WithEnvVars("foo=bar", "baz="),
			WithPrivileged(),
			WithWorkingDir("/foo"),
			WithUser("foo"),
			WithTTY(),
			WithConsoleSize(666, 42),
		)

		Expect(exopts.Conf.Env).To(ConsistOf("foo=bar", "baz="))
		Expect(exopts.Conf.Privileged).To(BeTrue())
		Expect(exopts.Conf.WorkingDir).To(Equal("/foo"))
		Expect(exopts.Conf.User).To(Equal("foo"))
		Expect(exopts.Conf.Tty).To(BeTrue())
		Expect(exopts.Conf.ConsoleSize).To(gstruct.PointTo(Equal([2]uint{42, 666})))
	})

})
