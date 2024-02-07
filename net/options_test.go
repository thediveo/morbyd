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

package net

import (
	"errors"

	"github.com/docker/docker/api/types/network"
	"github.com/thediveo/morbyd/ipam"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gs "github.com/onsi/gomega/gstruct"
)

var _ = Describe("network options", func() {

	It("processes net options", func() {
		netos := Options{}
		for _, opt := range []Opt{
			WithDriver("foobar"),
			WithIPAM(ipam.WithPool("0.0.1.0/24", ipam.WithRange("0.0.1.0/27"))),
			WithInternal(),
			WithIPv6(),
			WithLabel("foo=bar"),
			WithLabels("bar=baz"),
			WithOption("say=doh"),
		} {
			Expect(opt(&netos)).To(Succeed())
		}
		Expect(netos.Driver).To(Equal("foobar"))
		Expect(netos.IPAM).To(gs.PointTo(Equal(network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet:  "0.0.1.0/24",
					IPRange: "0.0.1.0/27",
				},
			}})))
		Expect(netos.Internal).To(BeTrue())
		Expect(netos.EnableIPv6).To(BeTrue())
		Expect(netos.Options).To(HaveLen(1))
		Expect(netos.Options).To(HaveKeyWithValue("say", "doh"))
		Expect(netos.Labels).To(HaveLen(2))
		Expect(netos.Labels).To(HaveKeyWithValue("foo", "bar"))
		Expect(netos.Labels).To(HaveKeyWithValue("bar", "baz"))
	})

	It("rejects invalid net options", func() {
		var opts Options
		Expect(WithLabels("=")(&opts)).To(HaveOccurred())
	})

	It("rejects invalid IPAM options", func() {
		failopt := func(*ipam.IPAM) error { return errors.New("error IJK305I") }
		var opts Options
		Expect(WithIPAM(failopt)(&opts)).To(HaveOccurred())
	})

})
