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
	"net"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("getting container IPs", Ordered, func() {

	It("returns a container's IP that we can talk to", func(ctx context.Context) {
		sess := Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=container.ip")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
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
		defer resp.Body.Close() //nolint:errcheck // any error is irrelevant at this point
		body := Successful(io.ReadAll(resp.Body))
		Expect(string(body)).To(Equal("Hellorld!\n"))
	})

	It("skips MACVLAN networks as not reachable from the host", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "NetworkInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		// These should be network IDs as opposed to names, but since these are
		// on par, using names makes our mocking more self-descriptive.
		rec.NetworkInspect(Any, mock.Eq("mac-wie-lahm"), Any).Return(network.Summary{
			Driver: "macvlan",
		}, nil)

		cntr := &Container{
			Session: sess,
			Details: container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"mac-wie-lahm": {
							NetworkID: "mac-wie-lahm",
							IPAddress: "1.0.1.1",
						},
					},
				},
			},
		}
		Expect(cntr.IP(ctx)).To(BeNil())
	})

	It("returns a nil IP in case of API errors", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "NetworkInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		// These should be network IDs as opposed to names, but since these are
		// on par, using names makes our mocking more self-descriptive.
		rec.NetworkInspect(Any, Any, Any).Return(network.Summary{}, errors.New("error IJK305I"))

		cntr := &Container{
			Session: sess,
			Details: container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"bridge-over-troubled-data": {
							NetworkID: "bridge-over-troubled-data",
							IPAddress: "1.0.2.1",
						},
					},
				},
			},
		}
		Expect(cntr.IP(ctx)).To(BeNil())
	})

	It("returns a nil IP in case of a none (null) network", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "NetworkInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		// These should be network IDs as opposed to names, but since these are
		// on par, using names makes our mocking more self-descriptive.
		rec.NetworkInspect(Any, mock.Eq("bubble"), Any).Return(network.Summary{
			Driver: "null",
		}, nil)
		rec.NetworkInspect(Any, mock.Eq("bridged"), Any).MinTimes(0).MaxTimes(1).Return(network.Summary{
			Driver: "bridge",
		}, nil)

		cntr := &Container{
			Session: sess,
			Details: container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"none": {
							NetworkID: "bubble",
							IPAddress: "",
						},
						"bridged": {
							NetworkID: "bridged",
							IPAddress: "",
						},
					},
				},
			},
		}
		Expect(cntr.IP(ctx)).To(BeNil())
	})

	It("returns loopback IP in case of a host network", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "NetworkInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		// These should be network IDs as opposed to names, but since these are
		// on par, using names makes our mocking more self-descriptive.
		rec.NetworkInspect(Any, mock.Eq("horscht"), Any).Return(network.Summary{
			Driver: "host",
		}, nil)

		cntr := &Container{
			Session: sess,
			Details: container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"host": {
							NetworkID: "horscht",
							IPAddress: "",
						},
					},
				},
			},
		}
		Expect(cntr.IP(ctx)).To(Equal(net.ParseIP("127.0.0.1")))
	})

	It("skips a network where we have no IP", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "NetworkInspect")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		// These should be network IDs as opposed to names, but since these are
		// on par, using names makes our mocking more self-descriptive.
		rec.NetworkInspect(Any, mock.Eq("bubble"), Any).Return(network.Summary{
			Driver: "bridge",
		}, nil)

		cntr := &Container{
			Session: sess,
			Details: container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"bubble": {
							NetworkID: "bubble",
							IPAddress: "",
						},
					},
				},
			},
		}
		Expect(cntr.IP(ctx)).To(BeNil())
	})

})
