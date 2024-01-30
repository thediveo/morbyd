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

package morbyd

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("run container", Ordered, func() {

	var sess *Session

	BeforeAll(func(ctx context.Context) {
		sess = Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
	})

	It("runs a container and captures its stdout and stderr", func(ctx context.Context) {
		msg := "D'OH!"
		errmsg := "D'OOOOHOOO!!"

		By("running a test container")
		var buff safe.Buffer
		var errbuf safe.Buffer

		// Providing an already primed input stream to the container results in
		// the script happily slurping in this data as soon as the container has
		// started. In turn, the script would already be finished and the
		// container already wound done by the time we want to inspect it for
		// its PID, causing spurious test failures. We thus send the script in
		// idle sleep that it'll be leaving only upon SIGTERM, where we control
		// when we send this SIGTERM.
		input := strings.NewReader(msg + "\n")
		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithCommand("/bin/sh", "-c",
				"trap \"exit\" SIGTERM; "+
					"IFS= read -r -s msg; echo \">>$msg<<\"; echo \""+errmsg+"\" 1>&2; "+
					"while true; do sleep 1s; done"),
			run.WithAutoRemove(),
			run.WithInput(input),
			// Using a timeline io.Writer helps with diagnosing potential test
			// issues, as we can relate the container's output to the test steps
			// as when these are happening.
			run.WithDemuxedOutput(io.MultiWriter(timestamper.New(GinkgoWriter), &buff),
				io.MultiWriter(timestamper.New(GinkgoWriter), &errbuf)),
		))
		Expect(cntr).NotTo(BeNil())
		defer func() {
			By("killing the container just to be sure (really, you cannot trust them otherwise)")
			cntr.Kill(ctx)
		}()

		By("retrieving the container's PID")
		Expect(cntr.PID(ctx)).NotTo(BeZero())

		By("talking to the container and receiving its answer")
		// Expect the correct, demux'ed output on stdout and stderr.
		Eventually(buff.String).Within(2 * time.Second).ProbeEvery(100 * time.Second).
			Should(Equal(">>" + msg + "<<\n"))
		Eventually(errbuf.String).Should(Equal(errmsg + "\n"))

		By("stopping the container")
		cntr.Stop(ctx)
	})

})
