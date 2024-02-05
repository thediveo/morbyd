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
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	types "github.com/docker/docker/api/types"
	"github.com/thediveo/morbyd/ipam"
	"github.com/thediveo/morbyd/net"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("containers", Ordered, func() {

	var sess *Session

	BeforeAll(func(ctx context.Context) {
		sess = Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
	})

	When("asking for a container's IP", func() {

		It("ignores MACVLAN IP addresses", func(ctx context.Context) {
			const testnetname = "morbyd-mcwielahm"

			netw := Successful(sess.CreateNetwork(ctx, testnetname,
				net.WithDriver("macvlan"),
				net.WithIPAM(ipam.WithPool("0.0.1.0/24"))))
			cntr := Successful(sess.Run(ctx, "busybox",
				run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; while true; do sleep 1; done"),
				run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
				run.WithNetwork(netw.ID)))
			Expect(cntr.IP(ctx)).To(BeNil())
		})

		It("returns a container's IP and we can talk to it", func(ctx context.Context) {
			if sess.IsDockerDesktop(ctx) {
				Skip("not on Docker Desktop")
			}

			const testnetname = "morbyd-bridge"
			netw := Successful(sess.CreateNetwork(ctx, testnetname))
			cntr := Successful(sess.Run(ctx, "busybox",
				run.WithCommand("/bin/sh", "-c",
					"mkdir /www; echo \"Hellorld!\" > /www/index.html; "+
						"httpd -v -f -h /www"),
				run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
				run.WithNetwork(netw.ID)))

			By("waiting for container initial process to have started")
			Expect(cntr.PID(ctx)).Error().NotTo(HaveOccurred())

			By("doing an HTTP exchange with the container")
			ip := cntr.IP(ctx)
			Expect(ip).NotTo(BeNil())
			get := Successful(http.NewRequestWithContext(ctx,
				http.MethodGet, fmt.Sprintf("http://%s/", ip), nil))
			clnt := &http.Client{Timeout: 5 * time.Second}
			resp := Successful(clnt.Do(get))
			defer resp.Body.Close()
			body := Successful(io.ReadAll(resp.Body))
			Expect(string(body)).To(Equal("Hellorld!\n"))
		})

	})

	It("waits for a container to finish", func(ctx context.Context) {
		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithCommand("/bin/sleep", "5s"),
			run.WithAutoRemove(),
			run.WithCombinedOutput(timestamper.New(GinkgoWriter)),
		))
		start := time.Now()
		Expect(cntr.Wait(ctx)).To(Succeed())
		Expect(time.Since(start)).To(BeNumerically(">=", 4*time.Second))
	})

	It("stops a container cooperatively", func(ctx context.Context) {
		var buff safe.Buffer

		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; echo \"OK\"; while true; do sleep 1; done"),
			run.WithAutoRemove(),
			run.WithCombinedOutput(io.MultiWriter(&buff, timestamper.New(GinkgoWriter))),
		))
		Eventually(buff.String).Within(5 * time.Second).ProbeEvery(100 * time.Millisecond).
			Should(ContainSubstring("OK"))
		cntr.Stop(ctx)
		Eventually(cntr.Refresh).WithContext(ctx).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
			Should(HaventFoundContainer())
	})

	It("kills a container without mercy", func(ctx context.Context) {
		var buff safe.Buffer

		cntr := Successful(sess.Run(ctx, "busybox",
			run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; echo \"OK\"; while true; do sleep 1; done"),
			run.WithCombinedOutput(io.MultiWriter(&buff, timestamper.New(GinkgoWriter))),
		))
		Eventually(buff.String).Within(5 * time.Second).ProbeEvery(100 * time.Millisecond).
			Should(ContainSubstring("OK"))
		cntr.Kill(ctx)
		Eventually(cntr.Refresh).WithContext(ctx).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
			Should(HaventFoundContainer())
	})

	It("returns an abbreviated container ID", func() {
		c := &Container{}
		Expect(c.AbbreviatedID()).To(Equal(""))

		hexdigits := "0123456789ABCDEF"
		id := make([]byte, 64)
		for idx := range id {
			id[idx] = hexdigits[rand.Intn(len(hexdigits))]
		}

		c = &Container{ID: string(id)}
		Expect(c.AbbreviatedID()).To(Equal(string(id)[:AbbreviatedIDLength]))
	})

	It("returns an error when container refresh fails", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ContainerInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{}, errors.New("error IJK305I"))

		cntr := &Container{Session: sess, ID: "bad1dea"}
		Expect(cntr.Refresh(ctx)).Error().To(MatchError(ContainSubstring("cannot refresh details of container")))
	})

	When("getting a container's PID", func() {

		It("retries until PID becomes available", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{},
			}, nil)
			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Pid: 42,
					},
				},
			}, nil)

			cntr := &Container{Session: sess, ID: "bad1dea"}
			Expect(cntr.PID(ctx)).To(Equal(42))
		})

		It("waits for a restart", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Dead:       true,
						Restarting: true,
					},
				},
			}, nil)
			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Pid: 42,
					},
				},
			}, nil)

			cntr := &Container{Session: sess, ID: "bad1dea"}
			Expect(cntr.PID(ctx)).To(Equal(42))
		})

		It("gives up when there's no chance", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ContainerInspect")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()

			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{},
			}, nil)
			rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						OOMKilled: true,
					},
				},
			}, nil)

			cntr := &Container{Session: sess, ID: "bad1dea"}
			Expect(cntr.PID(ctx)).Error().To(HaveOccurred())
		})

		When("handling API errors", func() {

			It("reports an error when container cannot be inspected", func(ctx context.Context) {
				ctrl := mock.NewController(GinkgoT())
				sess := Successful(NewSession(ctx,
					WithMockController(ctrl, "ContainerInspect")))
				DeferCleanup(func(ctx context.Context) {
					sess.Close(ctx)
				})
				rec := sess.Client().(*MockClient).EXPECT()

				rec.ContainerInspect(Any, Any).Return(types.ContainerJSON{}, errors.New("error IJK305I"))

				cntr := &Container{Session: sess, ID: "bad1dea"}
				Expect(cntr.PID(ctx)).Error().To(HaveOccurred())
			})

		})
	})

})
