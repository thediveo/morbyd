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
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("addresses", func() {

	Context("single address", func() {

		It("returns the transport protocol network the address is on", func() {
			Expect(Addr{}.Network()).To(BeEmpty())
			Expect(NewAddr(net.ParseIP("127.0.0.1"), 1234, "tcp").Network()).To(Equal("tcp"))
		})

		It("returns the stringified address", func() {
			Expect(Addr{}.String()).To(BeEmpty())
			Expect(NewAddr(net.ParseIP("127.0.0.1"), 1234, "tcp").String()).To(Equal("127.0.0.1:1234"))
			Expect(NewAddr(net.ParseIP("fe80::dead:beef"), 1234, "udp").String()).To(Equal("[fe80::dead:beef]:1234"))
		})

		It("uses loopback instead of unspecified", func() {
			Expect(Addr{}.UnspecifiedAsLoopback().String()).To(BeEmpty())
			Expect(NewAddr(net.ParseIP("127.0.0.1"), 1234, "tcp").UnspecifiedAsLoopback().String()).To(Equal("127.0.0.1:1234"))
			Expect(NewAddr(net.ParseIP("fe80::dead:beef"), 1234, "udp").UnspecifiedAsLoopback().String()).To(Equal("[fe80::dead:beef]:1234"))
			Expect(NewAddr(net.ParseIP("0.0.0.0"), 1234, "tcp").UnspecifiedAsLoopback().String()).To(Equal("127.0.0.1:1234"))
			Expect(NewAddr(net.ParseIP("::1"), 1234, "udp").UnspecifiedAsLoopback().String()).To(Equal("[::1]:1234"))
		})

	})

	Context("list of addresses", func() {

		It("returns zero addresses if list is empty", func() {
			Expect(Addrs{}.Any().String()).To(BeEmpty())
			Expect(Addrs{}.First().String()).To(BeEmpty())
		})

		It("returns always the first address element from a list", func() {
			addrs := Addrs{
				NewAddr(net.ParseIP("::"), 1234, "tcp"),
				NewAddr(net.ParseIP("0.0.0.0"), 1234, "tcp"),
				NewAddr(net.ParseIP("127.0.0.1"), 4567, "tcp"),
			}
			for i := 0; i < 10; i++ {
				Expect(addrs.First().String()).To(Equal("[::]:1234"))
			}
		})

		It("returns a random address element from a list", func() {
			addrs := Addrs{
				NewAddr(net.ParseIP("::"), 1234, "tcp"),
				NewAddr(net.ParseIP("0.0.0.0"), 1234, "tcp"),
				NewAddr(net.ParseIP("127.0.0.1"), 4567, "tcp"),
			}
			counts := map[string]int{}
			for i := 0; i < 100; i++ {
				a := addrs.Any().String()
				counts[a] = counts[a] + 1
			}
			Expect(counts).To(HaveLen(len(addrs)))
		})

	})

})
