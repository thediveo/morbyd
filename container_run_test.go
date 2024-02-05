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
	"errors"
	"io"
	"strings"
	"time"

	types "github.com/docker/docker/api/types"
	container "github.com/docker/docker/api/types/container"
	image "github.com/docker/docker/api/types/image"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"
	mock "go.uber.org/mock/gomock"

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

	When("failing", func() {

		It("reports failing options", func(ctx context.Context) {
			failopt := func(*run.Options) error { return errors.New("error IJK305I") }
			Expect(sess.Run(ctx, "", failopt)).Error().To(HaveOccurred())
		})

		It("reports failing image checks", func(ctx context.Context) {
			ctx, cancel := context.WithCancel(ctx)
			cancel()
			Expect(sess.Run(ctx, "busybox")).Error().To(MatchError(ContainSubstring("cannot run image")))
		})

		It("reports failing image pulls", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ImageList", "ImagePull")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ImageList(Any, Any).Return([]image.Summary{}, nil)
			rec.ImagePull(Any, Any, Any).Return(nil, errors.New("error IJK305I"))

			Expect(sess.Run(ctx, "busybox")).Error().To(MatchError(ContainSubstring("cannot pull image")))
		})

		It("reports creation failure", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerCreate")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerCreate(Any, Any, Any, Any, Any, Any).Return(container.CreateResponse{}, errors.New("error IJK305I"))

			Expect(sess.Run(ctx, "busybox")).Error().To(MatchError(ContainSubstring("cannot create container")))
		})

		It("reports attachment failure and cleans up", func(ctx context.Context) {
			const canaryName = "morbyd-canary-container"

			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				session.WithAutoCleaning("test.morbyd=container-run"),
				WithMockController(ctrl, "ContainerAttach")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerAttach(Any, Any, Any).Return(types.HijackedResponse{}, errors.New("error IJK305I"))

			Expect(sess.Run(ctx, "busybox", run.WithName(canaryName))).Error().To(MatchError(ContainSubstring("cannot attach to container")))
			Expect(sess.Container(ctx, canaryName)).Error().To(HaveOccurred())
		})

		It("reports start failure and cleans up", func(ctx context.Context) {
			const canaryName = "morbyd-canary-container"

			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				session.WithAutoCleaning("test.morbyd=container-run"),
				WithMockController(ctrl, "ContainerStart")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerStart(Any, Any, Any).Return(errors.New("error IJK305I"))

			Expect(sess.Run(ctx, "busybox", run.WithName(canaryName))).Error().To(MatchError(ContainSubstring("cannot start container")))
			Expect(sess.Container(ctx, canaryName)).Error().To(HaveOccurred())
		})

		It("reports update failure after start, and cleans up", func(ctx context.Context) {
			const canaryName = "morbyd-canary-container"

			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				session.WithAutoCleaning("test.morbyd=container-run"),
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{}, errors.New("error IJK305I"))

			Expect(sess.Run(ctx, "busybox", run.WithName(canaryName))).Error().To(MatchError(ContainSubstring("cannot inspect newly started container")))
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
