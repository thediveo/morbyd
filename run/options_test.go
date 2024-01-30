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
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
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
			HaveField("Conf.Tty", BeFalse()),
			HaveField("In", BeNil()),
			HaveField("Out", BeIdenticalTo(&stdout)),
			HaveField("Err", BeIdenticalTo(&stdout)),
		))

		Expect(opts(WithDemuxedOutput(&stdout, &stderr))).To(And(
			HaveField("Conf.Tty", BeFalse()),
			HaveField("In", BeNil()),
			HaveField("Out", BeIdenticalTo(&stdout)),
			HaveField("Err", BeIdenticalTo(&stderr)),
		))

		Expect(opts(WithInput(&stdin))).To(And(
			HaveField("Conf.Tty", BeFalse()),
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
			WithCustomInit(),
		)

		Expect(o.Name).To(Equal("loosing_lattice"))
		Expect(o.Conf.Cmd).To(ConsistOf("/bin/bash", "-c", "false"))
		Expect(o.Conf.Env).To(ConsistOf("foo=bar", "baz="))
		Expect(o.Conf.Labels).NotTo(HaveKey("hellorld"))
		Expect(o.Conf.Labels).To(And(
			HaveKeyWithValue("foo", "bar"),
			HaveKeyWithValue("baz", ""),
		))
		Expect(o.Conf.StopSignal).To(Equal("SIGDOOZE"))
		Expect(*o.Conf.StopTimeout).To(Equal(42))
		Expect(o.Conf.Tty).To(BeTrue())
		Expect(o.Host.Privileged).To(BeTrue())
		Expect(o.Host.CapAdd).To(ConsistOf("CAP_SUCCESS"))
		Expect(o.Host.CapDrop).To(ConsistOf("ALL"))
		Expect(o.Host.CgroupnsMode).To(Equal(container.CgroupnsMode("c-host")))
		Expect(o.Host.IpcMode).To(Equal(container.IpcMode("i-host")))
		Expect(o.Host.NetworkMode).To(Equal(container.NetworkMode("n-host")))
		Expect(o.Host.PidMode).To(Equal(container.PidMode("p-host")))
		Expect(o.Host.Tmpfs).To(HaveKeyWithValue("/tmp", ""))
		Expect(o.Host.Tmpfs).To(HaveKeyWithValue("/temp", "tmpfs-size=42"))
		Expect(o.Host.Devices).To(ConsistOf(
			container.DeviceMapping{PathOnHost: "/dev/foo", PathInContainer: "/dev/foo", CgroupPermissions: "rwm"},
			container.DeviceMapping{PathOnHost: "/dev/foo", PathInContainer: "/dev/fool", CgroupPermissions: "rwm"},
			container.DeviceMapping{PathOnHost: "/dev/foo", PathInContainer: "/dev/fool", CgroupPermissions: "r"},
		))
		Expect(o.Host.ReadonlyRootfs).To(BeTrue())
		Expect(o.Host.SecurityOpt).To(ConsistOf("all=unconfined"))
		Expect(o.Net.EndpointsConfig).To(And(
			HaveKeyWithValue("one", &network.EndpointSettings{NetworkID: "one"}),
			HaveKeyWithValue("two", &network.EndpointSettings{NetworkID: "two"}),
		))
		Expect(o.Host.ConsoleSize).To(Equal([2]uint{42, 666}))
		Expect(o.Conf.Hostname).To(Equal("foohost"))
		Expect(o.Host.RestartPolicy).To(Equal(container.RestartPolicy{
			Name:              "always",
			MaximumRetryCount: 666,
		}))
		Expect(o.Host.PublishAllPorts).To(BeTrue())
		Expect(o.Host.Init).To(gstruct.PointTo(BeTrue()))

		o = opts(WithLabel("foo=bar"))
		Expect(o.Conf.Labels).To(HaveKeyWithValue("foo", "bar"))

		o = Options{}
		Expect(WithLabels("=")(&o)).NotTo(Succeed())

		o = opts(WithTmpfsOpts("/temp", "tmpfs-size=42"))
		Expect(o.Host.Tmpfs).To(HaveKeyWithValue("/temp", "tmpfs-size=42"))
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

		Expect(o.Conf.Volumes).To(HaveLen(1))
		Expect(o.Conf.Volumes).To(HaveKey("/foo"))

		Expect(o.Host.Binds).To(ConsistOf(
			"/run:/run2:ro",
			"/fool:/bar:z",
			Successful(os.Getwd())+":/run",
		))

		Expect(o.Host.Mounts).To(ConsistOf(
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
		defer os.Chdir(cwd) //nolint:golint,errcheck
		tmpdir := Successful(os.MkdirTemp("", "on-my-way-out-*"))
		defer os.RemoveAll(tmpdir) //nolint:golint,errcheck
		Expect(os.Chdir(tmpdir)).To(Succeed())
		Expect(os.RemoveAll(tmpdir)).To(Succeed())
		Expect(bindVolumeToBind("./relative:/absolute")).To(Equal("./relative:/absolute"))
	})

	It("rejects when given a network in invalid long form", func() {
		var o Options
		Expect(WithNetwork("foo=bar")(&o)).NotTo(Succeed())
	})

})
