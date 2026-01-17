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
	"time"

	types "github.com/docker/docker/api/types"
	container "github.com/docker/docker/api/types/container"
	"github.com/thediveo/morbyd/exec"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/success"
)

var _ = Describe("execute command inside container", Ordered, func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Second).
			ShouldNot(HaveLeaked(goodgos))
	})

	When("executing commands", func() {

		var sess *Session

		BeforeEach(func(ctx context.Context) {
			sess = Successful(NewSession(ctx,
				session.WithAutoCleaning("test.morbyd=container.exec")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
		})

		It("rejects failing options", func(ctx context.Context) {
			failopt := func(*exec.Options) error {
				return errors.New("error IJK305I")
			}
			cntr := &Container{}
			Expect(cntr.Exec(ctx, exec.Command(""), failopt)).Error().To(HaveOccurred())
		})

		It("executes a command and waits for it to terminate", SpecTimeout(30*time.Second), func(ctx context.Context) {
			By("spinning up a test container")
			cntr := Successful(sess.Run(ctx, "busybox",
				run.WithCommand("/bin/sh", "-c",
					"trap 'echo \"test container cooperatively stopped\"; exit 1' TERM; echo \"test container started\"; "+
						"while true; do sleep 1; done; echo \"test container prematuredly stopped\""),
				run.WithAutoRemove(),
				run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
			))
			defer cntr.Stop(ctx)

			By("executing a command inside the container")
			exec := Successful(cntr.Exec(ctx,
				exec.Command("find", "/", "-samefile", "/bin/wgetxx"),
				exec.WithCombinedOutput(GinkgoWriter)))
			waitctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			Expect(exec.Wait(waitctx)).To(Equal(1))
		})

		It("executes a command using a TTY", SpecTimeout(30*time.Second), func(ctx context.Context) {
			By("spinning up a test container")
			cntr := Successful(sess.Run(ctx, "busybox",
				run.WithCommand("/bin/sh", "-c",
					"trap 'echo \"test container cooperatively stopped\"; exit 1' TERM; echo \"test container started\"; "+
						"while true; do sleep 1; done; echo \"test container prematuredly stopped\""),
				run.WithAutoRemove(),
				run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
			))
			defer cntr.Stop(ctx)

			By("executing a command inside the container")
			var buff safe.Buffer
			_ = Successful(cntr.Exec(ctx,
				exec.Command("/bin/sh", "-c", "echo \"**FOO!**\" 1>&2; while true; do sleep 1; done"),
				exec.WithTTY(),
				exec.WithCombinedOutput(io.MultiWriter(&buff, timestamper.New(GinkgoWriter)))))

			By("waiting for output")
			Eventually(buff.String).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				Should(Equal("**FOO!**\r\n"))

			cntr.Kill(ctx)
		})

		It("determines the executing command's PID", SpecTimeout(30*time.Second), func(ctx context.Context) {
			By("spinning up a test container")
			cntr := Successful(sess.Run(ctx, "busybox",
				run.WithCommand("/bin/sh", "-c",
					"trap 'echo \"test container cooperatively stopped\"; exit 1' TERM; echo \"test container started\"; "+
						"while true; do sleep 1; done; echo \"test container prematuredly stopped\""),
				run.WithAutoRemove(),
				run.WithCombinedOutput(GinkgoWriter),
			))
			defer func() {
				By("SIGTERM'ing the test container")
				cntr.Stop(ctx)
			}()

			By("executing a command inside the container")
			r, w := io.Pipe()
			defer w.Close() //nolint:errcheck // any error is irrelevant at this point
			defer r.Close() //nolint:errcheck // any error is irrelevant at this point
			execcmd := Successful(cntr.Exec(ctx,
				exec.Command("sh", "-c", "echo \"exec command started\"; read -s input; echo $input; echo \"exec command finished\"; exit 42"),
				exec.WithInput(r),
				exec.WithCombinedOutput(GinkgoWriter)))
			Expect(execcmd.Done()).NotTo(BeClosed())

			By("determining the PID of the executing command")
			Expect(execcmd.PID(ctx)).NotTo(BeZero())

			By("sending input to the executing command and closing writing end of input")
			Expect(w.Write([]byte("!\n"))).Error().NotTo(HaveOccurred())
			w.Close() //nolint:errcheck // any error is irrelevant at this point

			By("waiting for exit code")
			waitctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			Expect(execcmd.Wait(waitctx)).To(Equal(42))

			By("ensuring PID retrieval now returns an error")
			Expect(execcmd.PID(ctx)).Error().To(MatchError("command has already terminated"))
		})

		It("rejects non-existing container", func(ctx context.Context) {
			cntr := &Container{ID: "", Session: sess}
			Expect(cntr.Exec(ctx, exec.Command("foo"))).Error().To(HaveOccurred())
		})

	})

	When("the API fails", func() {

		It("reports when container-exec cannot be created", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect", "ContainerExecCreate")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(container.InspectResponse{}, nil)
			rec.ContainerExecCreate(Any, Any, Any).Return(container.ExecCreateResponse{}, errors.New("error IJK305I"))

			cntr := &Container{
				Session: sess,
			}
			Expect(cntr.Exec(ctx, exec.Command("foo"))).Error().To(HaveOccurred())
		})

		It("reports when attaching fails", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl,
					"ContainerInspect", "ContainerExecCreate", "ContainerExecAttach")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(container.InspectResponse{}, nil)
			rec.ContainerExecCreate(Any, Any, Any).Return(container.ExecCreateResponse{ID: "42"}, nil)
			rec.ContainerExecAttach(Any, Any, Any).Return(types.HijackedResponse{}, errors.New("error IJK305I"))

			cntr := &Container{
				Session: sess,
			}
			Expect(cntr.Exec(ctx, exec.Command("foo"))).Error().To(HaveOccurred())
		})

	})

})
