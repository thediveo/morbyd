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

package ipvlan

import (
	"github.com/thediveo/morbyd/net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ipvlan options", func() {

	It("processes ipvlan options", func() {
		nos := net.Options{}
		for _, opt := range []net.Opt{
			WithParent("pn666"),
			WithMode(L3Mode),
			WithFlag(BridgeFlag),
		} {
			Expect(opt(&nos)).To(Succeed())
		}
		Expect(nos.Options).To(HaveKeyWithValue("parent", "pn666"))
		Expect(nos.Options).To(HaveKeyWithValue("ipvlan_mode", "l3"))
		Expect(nos.Options).To(HaveKeyWithValue("ipvlan_flag", "bridge"))
	})

})
