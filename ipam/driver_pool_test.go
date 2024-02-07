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

package ipam

import (
	"errors"

	"github.com/docker/docker/api/types/network"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func WithPoolError() PoolOpt {
	return func(p *Pool) error { return errors.New("foobar!") }
}

var _ = Describe("ipam driver and pool options", func() {

	It("processes IPAM options, including Pool options", func() {
		ipamos := IPAM{}
		for _, opt := range []IPAMOpt{
			WithName("foobar"),
			WithPool("0.0.1.0/24", WithRange("0.0.1.0/27"),
				WithGateway("0.0.1.1"),
				WithAuxAddress("foo", "0.0.1.2"),
				WithAuxAddress("bar", "0.0.1.3")),
			WithPool("0.0.2.0/24"),
			WithOption("fool=barz"),
		} {
			Expect(opt(&ipamos)).To(Succeed())
		}
		Expect(ipamos.Driver).To(Equal("foobar"))
		Expect(ipamos.Config).To(ConsistOf(
			network.IPAMConfig{
				Subnet:  "0.0.1.0/24",
				IPRange: "0.0.1.0/27",
				Gateway: "0.0.1.1",
				AuxAddress: map[string]string{
					"foo": "0.0.1.2",
					"bar": "0.0.1.3",
				},
			},
			network.IPAMConfig{
				Subnet: "0.0.2.0/24",
			},
		))
		Expect(ipamos.Options).To(HaveLen(1))
		Expect(ipamos.Options).To(HaveKeyWithValue("fool", "barz"))

		Expect(WithOption("=")(&ipamos)).NotTo(Succeed())
		Expect(WithOption("blabla")(&ipamos)).NotTo(Succeed())
	})

	It("reports (IPAM) Pool errors", func() {
		var ipamopts IPAM
		Expect(WithPool("0.0.1.0/24", WithPoolError())(&ipamopts)).To(HaveOccurred())
	})

})
