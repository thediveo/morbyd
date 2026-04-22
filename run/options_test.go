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

package run

import (
	"bytes"
	"net/netip"
	"os"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"

	gs "github.com/onsi/gomega/gstruct"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

func opts(opts ...Opt) Options {
	GinkgoHelper()
	o := Options{}
	for _, opt := range opts {
		Expect(opt(&o)).To(Succeed())
	}
	return o
}

var _ = Describe("run (container) options", func() {

	It("processes input/output options", func() {
		var (
			stdin  strings.Reader
			stdout bytes.Buffer
			stderr bytes.Buffer
		)

		Expect(opts()).To(And(
			HaveField("In", BeNil()),
			HaveField("Out", BeNil()),
			HaveField("Err", BeNil()),
		))

		Expect(opts(WithCombinedOutput(&stdout))).To(And(
			HaveField("Opts.Config.Tty", BeFalse()),
			HaveField("In", BeNil()),
			HaveField("Out", BeIdenticalTo(&stdout)),
			HaveField("Err", BeIdenticalTo(&stdout)),
		))

		Expect(opts(WithDemuxedOutput(&stdout, &stderr))).To(And(
			HaveField("Opts.Config.Tty", BeFalse()),
			HaveField("In", BeNil()),
			HaveField("Out", BeIdenticalTo(&stdout)),
			HaveField("Err", BeIdenticalTo(&stderr)),
		))

		Expect(opts(WithInput(&stdin))).To(And(
			HaveField("Opts.Config.Tty", BeFalse()),
			HaveField("In", BeIdenticalTo(&stdin)),
			HaveField("Out", BeNil()),
			HaveField("Err", BeNil()),
		))
	})

	It("processes run options", func() {
		o := opts(
			WithName("loosing_lattice"),
			WithCommand("/bin/bash", "-c", "false"),
			WithEnvVars("foo=bar", "baz="),
			WithLabels("hellorld="),
			ClearLabels(),
			WithLabels("foo=bar", "baz="),
			WithStopSignal("SIGDOOZE"),
			WithStopTimeout(42),
			WithTTY(),
			WithAutoRemove(),
			WithPrivileged(),
			WithCapAdd("CAP_SUCCESS"),
			WithCapDropAll(),
			WithCgroupnsMode("c-host"),
			WithIPCMode("i-host"),
			WithNetworkMode("n-host"),
			WithPIDMode("p-host"),
			WithTmpfs("/tmp"),
			WithTmpfsOpts("/temp", "tmpfs-size=42"),
			WithDevice("/dev/foo"),
			WithDevice("/dev/foo:/dev/fool"),
			WithDevice("/dev/foo:/dev/fool:r"),
			WithReadOnlyRootfs(),
			WithSecurityOpt("all=unconfined"),
			WithNetwork("one"),
			WithNetwork("two"),
			WithConsoleSize(666, 42),
			WithHostname("foohost"),
			WithRestartPolicy("always", 666),
			WithAllPortsPublished(),
			WithPublishedPort("[fe80::dead:beef]:2345:1234"),
			WithPublishedPort("127.0.0.1:1234"),
			WithPublishedPort("127.0.0.2:12345/udp"),
			WithCPUSet("1,3,5,99"),
			WithMems("6,66"),
			WithCustomInit(),
		)

		Expect(o.Opts.Name).To(Equal("loosing_lattice"))
		Expect(o.Opts.Config.Cmd).To(ConsistOf("/bin/bash", "-c", "false"))
		Expect(o.Opts.Config.Env).To(ConsistOf("foo=bar", "baz="))
		Expect(o.Opts.Config.Labels).NotTo(HaveKey("hellorld"))
		Expect(o.Opts.Config.Labels).To(And(
			HaveKeyWithValue("foo", "bar"),
			HaveKeyWithValue("baz", ""),
		))
		Expect(o.Opts.Config.StopSignal).To(Equal("SIGDOOZE"))
		Expect(*o.Opts.Config.StopTimeout).To(Equal(42))
		Expect(o.Opts.Config.Tty).To(BeTrue())
		Expect(o.Opts.HostConfig.Privileged).To(BeTrue())
		Expect(o.Opts.HostConfig.CapAdd).To(ConsistOf("CAP_SUCCESS"))
		Expect(o.Opts.HostConfig.CapDrop).To(ConsistOf("ALL"))
		Expect(o.Opts.HostConfig.CgroupnsMode).To(Equal(container.CgroupnsMode("c-host")))
		Expect(o.Opts.HostConfig.IpcMode).To(Equal(container.IpcMode("i-host")))
		Expect(o.Opts.HostConfig.NetworkMode).To(Equal(container.NetworkMode("n-host")))
		Expect(o.Opts.HostConfig.PidMode).To(Equal(container.PidMode("p-host")))
		Expect(o.Opts.HostConfig.Tmpfs).To(HaveKeyWithValue("/tmp", ""))
		Expect(o.Opts.HostConfig.Tmpfs).To(HaveKeyWithValue("/temp", "tmpfs-size=42"))
		Expect(o.Opts.HostConfig.Devices).To(ConsistOf(
			container.DeviceMapping{PathOnHost: "/dev/foo", PathInContainer: "/dev/foo", CgroupPermissions: "rwm"},
			container.DeviceMapping{PathOnHost: "/dev/foo", PathInContainer: "/dev/fool", CgroupPermissions: "rwm"},
			container.DeviceMapping{PathOnHost: "/dev/foo", PathInContainer: "/dev/fool", CgroupPermissions: "r"},
		))
		Expect(o.Opts.HostConfig.ReadonlyRootfs).To(BeTrue())
		Expect(o.Opts.HostConfig.SecurityOpt).To(ConsistOf("all=unconfined"))
		Expect(o.Opts.NetworkingConfig.EndpointsConfig).To(And(
			HaveKeyWithValue("one", &network.EndpointSettings{NetworkID: "one"}),
			HaveKeyWithValue("two", &network.EndpointSettings{NetworkID: "two"}),
		))
		Expect(o.Opts.HostConfig.ConsoleSize).To(Equal([2]uint{42, 666}))
		Expect(o.Opts.Config.Hostname).To(Equal("foohost"))
		Expect(o.Opts.HostConfig.RestartPolicy).To(Equal(container.RestartPolicy{
			Name:              "always",
			MaximumRetryCount: 666,
		}))

		Expect(o.Opts.HostConfig.PublishAllPorts).To(BeTrue())
		Expect(o.Opts.Config.ExposedPorts).To(HaveLen(2))
		Expect(o.Opts.Config.ExposedPorts).To(HaveKey(network.MustParsePort("1234/tcp")))
		Expect(o.Opts.Config.ExposedPorts).To(HaveKey(network.MustParsePort("12345/udp")))
		Expect(o.Opts.HostConfig.PortBindings).To(HaveLen(2))
		Expect(o.Opts.HostConfig.PortBindings).To(HaveKeyWithValue(
			network.MustParsePort("1234/tcp"),
			ConsistOf(
				network.PortBinding{HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: "0"},
				network.PortBinding{HostIP: netip.MustParseAddr("fe80::dead:beef"), HostPort: "2345"})))
		Expect(o.Opts.HostConfig.PortBindings).To(HaveKeyWithValue(
			network.MustParsePort("12345/udp"),
			ConsistOf(network.PortBinding{HostIP: netip.MustParseAddr("127.0.0.2"), HostPort: "0"})))

		Expect(o.Opts.HostConfig.CpusetCpus).To(Equal("1,3,5,99"))
		Expect(o.Opts.HostConfig.CpusetMems).To(Equal("6,66"))
		Expect(o.Opts.HostConfig.Init).To(gs.PointTo(BeTrue()))

		o = opts(WithLabel("foo=bar"))
		Expect(o.Opts.Config.Labels).To(HaveKeyWithValue("foo", "bar"))

		o = Options{}
		Expect(WithLabels("=")(&o)).NotTo(Succeed())

		o = opts(WithTmpfsOpts("/temp", "tmpfs-size=42"))
		Expect(o.Opts.HostConfig.Tmpfs).To(HaveKeyWithValue("/temp", "tmpfs-size=42"))
	})

	It("rejects invalid published port mappings", func() {
		var o Options
		Expect(WithPublishedPort("abcd")(&o)).To(HaveOccurred())
	})

	It("rejects invalid volume specs", func() {
		var o Options
		Expect(WithVolume("rappel:zappel:humba:tätärä")(&o)).To(
			MatchError("malformed WithVolume parameter \"rappel:zappel:humba:tätärä\", reason: invalid spec: rappel:zappel:humba:tätärä: too many colons"))
	})

	It("rejects invalid mount specs", func() {
		var o Options
		Expect(WithMount("type=bind,source=/foo,target=/bar,private")(&o)).To(
			MatchError("invalid WithMount parameter, reason: invalid field 'private' must be a key=value pair"))
	})

	It("rejects invalid devices", func() {
		var o Options
		Expect(WithDevice(":::::")(&o)).Error().To(
			MatchError(ContainSubstring("malformed WithDevice parameter")))
		Expect(WithDevice("::")(&o)).Error().To(
			MatchError("WithDevice host path parameter must not be empty"))
	})

	It("splits into volumes, binds, and mounts", func() {
		o := opts(
			WithVolume("/foo"),
			WithVolume("/run:/run2:ro"),
			WithVolume("/fool:/bar:z"),
			WithVolume(".:/run"),
			WithMount("type=volume,source=/foo,target=/bar,readonly"),
		)

		Expect(o.Opts.Config.Volumes).To(HaveLen(1))
		Expect(o.Opts.Config.Volumes).To(HaveKey("/foo"))

		Expect(o.Opts.HostConfig.Binds).To(ConsistOf(
			"/run:/run2:ro",
			"/fool:/bar:z",
			Successful(os.Getwd())+":/run",
		))

		Expect(o.Opts.HostConfig.Mounts).To(ConsistOf(
			mount.Mount{
				Type:     "volume",
				Source:   "/foo",
				Target:   "/bar",
				ReadOnly: true,
			},
		))
	})

	It("returns invalid volume string when converting to binds unmodified", Serial, func() {
		Expect(bindVolumeToBind("")).To(BeEmpty())

		cwd := Successful(os.Getwd())
		defer os.Chdir(cwd) //nolint:errcheck // any error is irrelevant at this point
		tmpdir := Successful(os.MkdirTemp("", "on-my-way-out-*"))
		defer os.RemoveAll(tmpdir) //nolint:errcheck // any error is irrelevant at this point
		Expect(os.Chdir(tmpdir)).To(Succeed())
		Expect(os.RemoveAll(tmpdir)).To(Succeed())
		Expect(bindVolumeToBind("./relative:/absolute")).To(Equal("./relative:/absolute"))
	})

	It("rejects when given a network in invalid long form", func() {
		var o Options
		Expect(WithNetwork("foo=bar")(&o)).NotTo(Succeed())
	})

	DescribeTable("published port mapping syntax",
		func(mapping string, expectedIP netip.Addr, expectedHostPort int, expectedCntrPort int, expectedL4Proto string, ok bool) {
			ip, hp, cp, l4p, err := parsePortMapping(mapping)
			if !ok {
				Expect(err).To(HaveOccurred())
				Expect(ip.IsValid()).To(BeFalse())
				Expect(hp).To(BeZero())
				Expect(cp).To(BeZero())
				Expect(l4p).To(BeEmpty())
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(ip).To(Equal(expectedIP))
			Expect(hp).To(Equal(uint16(expectedHostPort)))
			Expect(cp).To(Equal(uint16(expectedCntrPort)))
			Expect(l4p).To(Equal(expectedL4Proto))
		},

		Entry(nil, "", nil, 0, 0, "", false),

		// Nope
		Entry(nil, "0", nil, 0, 0, "", false),
		Entry(nil, "123abc", nil, 0, 0, "", false),
		Entry(nil, "2345:123abc", nil, 0, 0, "", false),
		Entry(nil, "2345xyz:123abc", nil, 0, 0, "", false),

		// Everything wrong with a potential IPv6 host address...
		Entry(nil, "[1234:", nil, 0, 0, "", false),
		Entry(nil, "[1234]", nil, 0, 0, "", false),
		Entry(nil, "[1234]:", nil, 0, 0, "", false),

		// Aisle of Plenty :D
		Entry(nil, "[::1]:2345:1234:7890", nil, 0, 0, "", false),
		Entry(nil, "127.0.0.1:2345:1234:7890", nil, 0, 0, "", false),

		// Oddballs, odd, yet fine.
		Entry(nil, "1234/tcp/udp", nil, 0, 1234, "tcp/udp", true),

		// Good cases.
		Entry(nil, "1234", nil, 0, 1234, "tcp", true),
		Entry(nil, "1234/udp", nil, 0, 1234, "udp", true),

		Entry(nil, "127.0.0.1:1234", netip.MustParseAddr("127.0.0.1"), 0, 1234, "tcp", true),
		Entry(nil, "[::1]:1234", netip.MustParseAddr("::1"), 0, 1234, "tcp", true),

		Entry(nil, "2345:1234", nil, 2345, 1234, "tcp", true),
		Entry(nil, "127.0.0.1:2345:1234", netip.MustParseAddr("127.0.0.1"), 2345, 1234, "tcp", true),
		Entry(nil, "[::1]:2345:1234", netip.MustParseAddr("::1"), 2345, 1234, "tcp", true),
	)

	DescribeTable("user (with group) principals, not principles",
		func(p any, expected string) {
			var o Options
			switch v := p.(type) {
			case string:
				_ = WithUser(v)(&o)
			case int:
				_ = WithUser(v)(&o)
			}
			Expect(o.Opts.Config.User).To(Equal(expected))
		},
		Entry("clear", "", ""),
		Entry("user name", "maroding_marble", "maroding_marble"),
		Entry("user name", 1000, "1000"),
		Entry("user:group names", "peter:paul-marry", "peter:paul-marry"),
	)

	DescribeTable("user/group principals",
		func(pu string, pg any, expected string) {
			var o Options
			_ = WithUser(pu)(&o)
			switch v := pg.(type) {
			case string:
				_ = WithGroup(v)(&o)
			case int:
				_ = WithGroup(v)(&o)
			}
			Expect(o.Opts.Config.User).To(Equal(expected))
		},
		Entry("clear group", "1000:6666", "", "1000"),
		Entry("replace group", "1000:6666", "blue_breach", "1000:blue_breach"),
		Entry("replace group (int)", "1000:6666", 42, "1000:42"),
		Entry("add group", "1000", "blue_breach", "1000:blue_breach"),
	)
})
