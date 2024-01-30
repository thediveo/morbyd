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

package bridge

import (
	"github.com/thediveo/morbyd/net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("bridge options", func() {

	It("processes bridge options", func() {
		nos := net.Options{}
		for _, opt := range []net.Opt{
			WithBridgeName("brrrr"),
			WithoutIPMasquerade(),
			WithoutICC(),
			WithMTU(666),
			WithInterfacePrefix("brrrrr"),
		} {
			Expect(opt(&nos)).To(Succeed())
		}
		Expect(nos.Options).To(HaveKeyWithValue("com.docker.network.bridge.name", "brrrr"))
		Expect(nos.Options).To(HaveKeyWithValue("com.docker.network.bridge.enable_ip_masquerade", "false"))
		Expect(nos.Options).To(HaveKeyWithValue("com.docker.network.bridge.enable_icc", "false"))
		Expect(nos.Options).To(HaveKeyWithValue("com.docker.network.driver.mtu", "666"))
		Expect(nos.Options).To(HaveKeyWithValue("com.docker.network.container_iface_prefix", "brrrrr"))
	})

})
