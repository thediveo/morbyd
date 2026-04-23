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
	"fmt"
	"io"
	"net"
	"net/netip"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"

	"github.com/thediveo/morbyd/v2/identity"
	"github.com/thediveo/morbyd/v2/internal/ensure"
	lbls "github.com/thediveo/morbyd/v2/labels"
	"github.com/thediveo/morbyd/v2/run/dockercli"
	"github.com/thediveo/morbyd/v2/strukt"
)

// Opt is a configuration option to run a container using
// [github.com/thediveo/morbyd/v2.Session.Run]. Please see also [Options] for
// more information.
type Opt func(*Options) error

// Options represent the plethora of options for creating a container from an
// image, attaching to it, and finally starting it, as well as additional
// options for handling the input and output of containers.
//
// Please note that the defaults are:
//   - don't allocate a pseudo TTY (a.k.a “-t” CLI arg); that means that the
//     container gets fifos/pipes assigned to its stdin, stdout, and stderr,
//     but not a pseudo TTY.
//   - don't allocate stdin; use [WithInput] to set an [io.Reader] from which
//     the container gets its stdin data stream fed. Please note that this
//     additionally also sets OpenStdIn and StdinOnce. However, [WithInput]
//     is not the same as the “-i” CLI arg; “-i” additionally sets AttachStdin,
//     an API option that still is unknown to as what exactly it does during
//     container creation...
type Options struct {
	Opts client.ContainerCreateOptions
	In   io.Reader
	Out  io.Writer
	Err  io.Writer
}

// WithCombinedOutput sends the container's stdout and stderr to the specified
// io.Writer. This also automatically attaches the container's stdout and stderr
// after the container has been created.
//
// Please note that you should WithCombinedOutput when using [WithTTYP],as
// Docker then combines the container's stdout and stderr into a single output
// stream.
func WithCombinedOutput(w io.Writer) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Out = w
		o.Err = w
		return nil
	}
}

// WithDemuxedOutput disables the TTY option in order to send the container's
// stdout and stderr properly separated to the specified out and err io.Writer.
//
// It is a known limitation of the (pseudo) TTY to combine both stdout and
// stderr of the container's output streams, and there is no way for Docker (and
// thus for us) to demultiplex this output sludge after the fact.
func WithDemuxedOutput(out io.Writer, err io.Writer) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.Tty = false
		o.Out = out
		o.Err = err
		return nil
	}
}

// WithInput sends input data from the specified io.Reader to the container's
// stdin. For this, it allocates a stdin for the container when creating it and
// then attaches to this stdin after creation.
func WithInput(r io.Reader) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.OpenStdin = true
		o.Opts.Config.StdinOnce = true
		o.In = r
		return nil
	}
}

// WithName sets the optional name of the container to create.
func WithName(name string) Opt {
	return func(o *Options) error {
		o.Opts.Name = name
		return nil
	}
}

// WithCommand sets the optional command to execute at container start.
func WithCommand(cmd ...string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.Cmd = cmd
		return nil
	}
}

// WithEnvVars adds multiple environment variables to the container to be
// started.
func WithEnvVars(vars ...string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.Env = append(o.Opts.Config.Env, vars...)
		return nil
	}
}

// ClearLabels clears any labels inherited from the [Session] when creating a
// new container.
func ClearLabels() Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		clear(o.Opts.Config.Labels)
		return nil
	}
}

// WithLabel adds a label in “key=value” format to the container's labels.
func WithLabel(label string) Opt {
	return func(o *Options) error {
		ensureLabelsMap(o)
		return lbls.Labels(o.Opts.Config.Labels).Add(label)
	}
}

// WithLabels adds multiple labels in “key=value” format to the container's
// labels.
func WithLabels(labels ...string) Opt {
	return func(o *Options) error {
		ensureLabelsMap(o)
		for _, label := range labels {
			if err := lbls.Labels(o.Opts.Config.Labels).Add(label); err != nil {
				return err
			}
		}
		return nil
	}
}

func ensureLabelsMap(o *Options) {
	ensure.Value(&o.Opts.Config)
	ensure.Map(&o.Opts.Config.Labels)
}

// WithStopSignal sets the name of the signal to be sent to its initial process
// when stopping the container.
func WithStopSignal(s string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.StopSignal = s
		return nil
	}
}

// WithStopTimeout sets the timeout (in seconds) to stop the container.
func WithStopTimeout(secs int) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.StopTimeout = &secs
		return nil
	}
}

// WithTTY allocates a pseudo TTY for the container's input and output.
//
// Please note that using a TTY causes the container's stdout and stderr streams
// to get mixed together, so this option can be used only in combination with
// [WithCombinedOutput]. Specifying it after [WithDemuxedOutput] will cause the
// combined stdout+stderr output to appear only on the specified output
// io.Writer, whereas the specified error io.Writer won't receive any (error)
// output at all.
//
// When WithTTY is used before [WithDemuxedOutput], it'll become ineffective.
func WithTTY() Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.Tty = true
		return nil
	}
}

// WithAutoRemove removes the container after it has stopped.
func WithAutoRemove() Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.AutoRemove = true
		return nil
	}
}

// WithPrivileged runs the container as privileged, including all capabilities
// and not masking out certain filesystem elements.
func WithPrivileged() Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.Privileged = true
		return nil
	}
}

// WithCapAdd adds an individual kernel capabilities to the initial process in
// the container.
func WithCapAdd(cap string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.CapAdd = append(o.Opts.HostConfig.CapAdd, cap)
		return nil
	}
}

// WithCapDropAll drops all kernel capabilities for the initial process in the
// container.
//
// Please note that we don't provide dropping individual capabilities on
// purpose: the default set of Docker container capabilities depend on the
// Docker engine and Linux kernel, so this would result in non-deterministic
// behavior. Drop them all, or drop none.
func WithCapDropAll() Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.CapDrop = []string{"ALL"}
		return nil
	}
}

// WithCgroupnsMode configures the cgroup namespace to use when creating the new
// container; it can be either “” (empty, use the daemon's default)
// configuration, “private”, or “host”.
func WithCgroupnsMode(mode string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.CgroupnsMode = container.CgroupnsMode(mode)
		return nil
	}
}

// WithIPCMode configures the IPC namespace to use when creating the new
// container; it can be either “” (empty, use the daemon's default), “none”
// (private, but without /dev/shm mounted), “private”, “shareable”, “host”, or
// container:NAMEID”.
func WithIPCMode(mode string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.IpcMode = container.IpcMode(mode)
		return nil
	}
}

// WithNetworkMode configures the net(work) namespace to use when creating the new
// container; it can be either “” (create a new one), “none”, “host”, or
// container:NAMEID”.
func WithNetworkMode(mode string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.NetworkMode = container.NetworkMode(mode)
		return nil
	}
}

// WithPIDMode configures the PID namespace to use when creating the new
// container; it can be either “” (create a new one), “host”, or
// container:NAMEID”.
func WithPIDMode(mode string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.PidMode = container.PidMode(mode)
		return nil
	}
}

// WithPublishedPort exposes a container's port on the host, similar to the “-p”
// and “--publish” flags in the “docker create” and “docker run” CLI commands.
// The mapping syntax supported is a superset of Docker's, supporting a random
// host port while binding to a specific host IP address only:
//
//	[HOSTIP:][HOSTPORT:]CONTAINERPORT[/L4PROTO]
//
// An IPv6 HOSTIP needs to be in “[::]” format.
//
// Illustrative examples:
//
//   - "1234" publishes the container's TCP port 1234 on a random, available host
//     TCP port, bound to the host's unspecified IP address(es).
//   - "1234/tcp" is the same as "1234".
//   - "1234:1234" publishes the container's TCP port 1234 on the host's TCP port
//     1234, bound to the host's unspecified IP address(es).
//   - (superset) "127.0.0.1:1234" publishes the container's TCP port 1234 on a
//     random, available host TCP port, bound the the IPv4 loopback address
//     127.0.0.1.
//   - (superset) "127.0.0.1:1234/tcp" is the same as "127.0.0.1:1234".
//   - "127.0.0.1:666:1234" publishes the container's TCP port 1234 on the host's
//     TCP port 666, bound to the IPv4 loopback address 127.0.0.1.
//   - "[::1]:1234" publishes the container's TCP port 1234 on a random, available
//     host TCP port, bound the the host's IPv6 loopback address ::1.
func WithPublishedPort(mapping string) Opt {
	return func(o *Options) error {
		bindIP, hostPort, cntrPort, l4proto, err := parsePortMapping(mapping)
		if err != nil {
			return err
		}
		ensure.Value(&o.Opts.Config)
		ensure.Value(&o.Opts.HostConfig)
		// ouch, we need to set also ExposedPorts, otherwise the PortBindings
		// get ignored.
		if o.Opts.Config.ExposedPorts == nil {
			o.Opts.Config.ExposedPorts = network.PortSet{}
		}
		if o.Opts.HostConfig.PortBindings == nil {
			o.Opts.HostConfig.PortBindings = network.PortMap{}
		}
		portProtoText := fmt.Sprintf("%d/%s", cntrPort, l4proto)
		portProto, err := network.ParsePort(portProtoText)
		if err != nil {
			return err
		}
		o.Opts.Config.ExposedPorts[portProto] = struct{}{}
		o.Opts.HostConfig.PortBindings[portProto] = append(o.Opts.HostConfig.PortBindings[portProto], network.PortBinding{
			HostIP:   bindIP,
			HostPort: strconv.FormatUint(uint64(hostPort), 10),
		})
		return nil
	}
}

func parsePortMapping(mapping string) (bindIP netip.Addr, hostPort uint16, cntrPort uint16, l4proto string, err error) {
	// Split off the optional transport protocol protocol name and default
	// to "tcp" if not specified.
	addrsports, l4proto, _ := strings.Cut(mapping, "/")
	if l4proto == "" {
		l4proto = "tcp"
	}
	// Strip off the IPv6 address, if present, before we proceed to chop the
	// mapping into pieces.
	hasBindIP := false
	if strings.HasPrefix(addrsports, "[") {
		var ip string
		var ok bool
		ip, addrsports, ok = strings.Cut(addrsports, "]:")
		if !ok {
			return reportPortMappingError(mapping)
		}
		bindIP, err = netip.ParseAddr(ip[1:])
		if err != nil {
			return reportPortMappingError(mapping)
		}
		hasBindIP = true
	}
	// Now chop the mapping into at most three fields, or at most two fields
	// we had already just chopped off the IPv6 address.
	fields := strings.Split(addrsports, ":")
	if len(fields) > 3 || (hasBindIP && len(fields) > 2) {
		return reportPortMappingError(mapping)
	}
	// Does the first field look like an IPv4 address in case we hadn't already
	// an IPv6 address?
	if !hasBindIP {
		bindIP, err = netip.ParseAddr(fields[0])
		if err == nil {
			fields = fields[1:]
		}
	}
	if len(fields) == 2 {
		hp, err := strconv.ParseUint(fields[0], 10, 16)
		if err != nil {
			return reportPortMappingError(mapping)
		}
		hostPort = uint16(hp)
		fields = fields[1:]
	}
	cp, err := strconv.ParseUint(fields[0], 10, 16)
	if err != nil || cp == 0 {
		return reportPortMappingError(mapping)
	}
	cntrPort = uint16(cp)
	return // sic!
}

func reportPortMappingError(mapping string) (bindIP netip.Addr, hostPort uint16, cntrPort uint16, l4proto string, err error) {
	return netip.Addr{}, 0, 0, "", fmt.Errorf("invalid port publishing mapping, expected [HOSTIP:][HOSTPORT:]CONTAINERPORT[/L4PROTO], got: %s",
		mapping)
}

// WithVolume adds a volume, in the format “source:target:options”. Please
// see also Docker's [Volumes] documentation.
//
// - “/var”: an anonymous volume.
//
// Options are comma-separated values:
//   - “ro” for read-only; if not present, the default is read-write.
//   - “z” (sharing content among multiple containers) or “Z” (content
//     is private and unshared).
//
// [Volumes]: https://docs.docker.com/storage/volumes/
func WithVolume(vol string) Opt {
	parsedVol, err := dockercli.ParseVolume(vol)
	if err != nil {
		return func(o *Options) error {
			return fmt.Errorf("malformed WithVolume parameter %q, reason: %w",
				vol, err)
		}
	}

	if parsedVol.Source == "" {
		// Anonymous volumes are still handled via the Volumes configuration API
		// field.
		return func(o *Options) error {
			ensure.Value(&o.Opts.Config)
			if o.Opts.Config.Volumes == nil {
				o.Opts.Config.Volumes = map[string]struct{}{}
			}
			o.Opts.Config.Volumes[vol] = struct{}{}
			return nil
		}
	}

	// All other volume specs are now handled via the host configuration bind
	// (mounts).
	if parsedVol.Type == string(mount.TypeBind) {
		vol = bindVolumeToBind(vol)
	}
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.Binds = append(o.Opts.HostConfig.Binds, vol)
		return nil
	}
}

// bindVolumeToBind converts a volume string known to actually be of type bind
// to its corresponding bind type string.
func bindVolumeToBind(vol string) string {
	source, target, ok := strings.Cut(vol, ":")
	if !ok {
		return vol
	}
	if source != "." && !strings.HasPrefix(source, "./") {
		return vol
	}
	abssrc, err := filepath.Abs(source)
	if err != nil {
		return vol
	}
	return abssrc + ":" + target
}

// WithMount adds a (bind) mount. Please see also Docker's [Bind mounts]
// documentation.
//
// [Bind mounts]: https://docs.docker.com/storage/bind-mounts/
func WithMount(mnt string) Opt {
	return func(o *Options) error {
		// Let's do a dry run on this single parameter first...
		dry := dockercli.MountOpt{}
		if err := dry.Set(mnt); err != nil { // ...actually an "add"
			return fmt.Errorf("invalid WithMount parameter, reason: %w",
				err)
		}
		mount := dry.Value()[0]
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.Mounts = append(o.Opts.HostConfig.Mounts, mount)
		return nil
	}
}

// WithTmpfs specifies the path inside the container to mount a new tmpfs
// instance on, using default options (unlimited size, world-writable).
func WithTmpfs(path string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		if o.Opts.HostConfig.Tmpfs == nil {
			o.Opts.HostConfig.Tmpfs = map[string]string{}
		}
		o.Opts.HostConfig.Tmpfs[path] = ""
		return nil
	}
}

// WithTmpfsOpts specifies the path isnide the container to mount a new tmpfs
// instance, as well as options in form of comma-separated “key=value”.
//
// These options are available:
//   - tmpfs-size=<bytes>; defaults to unlimited.
//   - tmpfs-mode=<oct>, where <oct> format can be “700” and “0770”. Defaults
//     to “1777”, that is, “world-writable”.
func WithTmpfsOpts(path string, opts string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		if o.Opts.HostConfig.Tmpfs == nil {
			o.Opts.HostConfig.Tmpfs = map[string]string{}
		}
		o.Opts.HostConfig.Tmpfs[path] = opts
		return nil
	}
}

// WithDevice specifies a host device to be added to the container to be
// created. The format is either:
//   - /dev/foo
//   - /dev/foo:/dev/bar
//   - /dev/foo:/dev/bar:rwm
//
// WithDevice will panic if the first (host path) element is empty. If the
// container path is empty, the same path as in the host is assumed.The cgroup
// permissions default to “rwm”.
//
// See also:
// https://docs.docker.com/engine/reference/commandline/container_run/#device
func WithDevice(dev string) Opt {
	return func(o *Options) error {
		var device struct {
			HostPath      string
			ContainerPath string
			CgroupPerms   string
		}
		if err := strukt.Unmarshal(dev, ":", &device); err != nil {
			return fmt.Errorf("malformed WithDevice parameter %q, reason: %w",
				dev, err)
		}
		if device.HostPath == "" {
			return fmt.Errorf("WithDevice host path parameter must not be empty")
		}
		if device.ContainerPath == "" {
			device.ContainerPath = device.HostPath
		}
		if device.CgroupPerms == "" {
			device.CgroupPerms = "rwm" // read-write-mknod
		}
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.Devices = append(o.Opts.HostConfig.Devices, container.DeviceMapping{
			PathOnHost:        device.HostPath,
			PathInContainer:   device.ContainerPath,
			CgroupPermissions: device.CgroupPerms,
		})
		return nil
	}
}

// WithReadOnlyRootfs configures the new container to use a ready-only topmost
// layer.
func WithReadOnlyRootfs() Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.ReadonlyRootfs = true
		return nil
	}
}

// WithSecurityOpt configures an additional security option, such as
// “seccomp=unconfined”.
func WithSecurityOpt(opt string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.SecurityOpt = append(o.Opts.HostConfig.SecurityOpt, opt)
		return nil
	}
}

// WithConsoleSize sets with width and height of the pseudo TTY; please note the
// width-height order in contrast to the Docker API order.
func WithConsoleSize(width, height uint) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.ConsoleSize = [2]uint{height, width} // sic!
		return nil
	}
}

// WithNetwork attaches the new container to a particular network. The nameid
// parameter either identifies a network by its name or ID, or can be in long
// format, consisting of comma-separated key-value pairs.
//
//   - “bridge”
//   - “name=bridge,ip=128.0.0.1”
//
// Please do not confuse with [WithNetworkMode] mode, where the latter
// configures the Linux kernel net namespace to use.
func WithNetwork(netw string) Opt {
	return func(o *Options) error {
		dry := dockercli.NetworkOpt{}
		if err := dry.Set(netw); err != nil {
			return fmt.Errorf("malformed WithNetwork parameter %q, reason: %s",
				netw, err)
		}
		ep := dry.Value()[0]
		ensure.Value(&o.Opts.NetworkingConfig)
		if o.Opts.NetworkingConfig.EndpointsConfig == nil {
			o.Opts.NetworkingConfig.EndpointsConfig = map[string]*network.EndpointSettings{}
		}
		var err error
		var ipv4, ipv6 netip.Addr
		if ep.IPv4Address != "" {
			if ipv4, err = netip.ParseAddr(ep.IPv4Address); err != nil {
				return fmt.Errorf("invalid IP address, reason: %s", err)
			}
		}
		if ep.IPv6Address != "" {
			if ipv6, err = netip.ParseAddr(ep.IPv6Address); err != nil {
				return fmt.Errorf("invalid IPv6 address, reason: %s", err)
			}
		}
		var mac net.HardwareAddr
		if ep.MacAddress != "" {
			if mac, err = net.ParseMAC(ep.MacAddress); err != nil {
				return fmt.Errorf("invalid MAC address, reason: %s", err)
			}
		}
		o.Opts.NetworkingConfig.EndpointsConfig[ep.Target] = &network.EndpointSettings{
			NetworkID:         ep.Target,
			Aliases:           ep.Aliases,
			DriverOpts:        ep.DriverOpts,
			Links:             ep.Links,
			IPAddress:         ipv4,
			GlobalIPv6Address: ipv6, // TODO: clarify w. respect to GlobalIPv6PrefixLen
			MacAddress:        network.HardwareAddr(mac),
		}
		return nil
	}
}

// WithHostname configures the host name to use inside the container.
func WithHostname(host string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.Hostname = host
		return nil
	}
}

// WithRestartPolicy configures the restart policy (“no”, “always”,
// “on-failure”, “unless-stopped”) as well as the maximum attempts at restarting
// the container.
func WithRestartPolicy(policy string, maxretry int) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.RestartPolicy = container.RestartPolicy{
			Name:              container.RestartPolicyMode(policy),
			MaximumRetryCount: maxretry,
		}
		return nil
	}
}

// WithAllPortsPublished instructs Docker to publish all exposed ports on the
// host. Use with great care.
func WithAllPortsPublished() Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.PublishAllPorts = true
		return nil
	}
}

// WithCPUSet configures the [set of CPUs] on which processes of the container
// are allowed to execute. The list of CPUs is a comma-separated list of CPU
// numbers and ranges of numbers, where the numbers are decimal numbers. For
// instance, “1,3,5”, "0-42,666", et cetera.
//
// To avoid stuttering this option is simply named “WithCPUSet” instead of
// “WithCpusetCPUs”, or similar awkward letter salads.
//
// [set of CPUs]: https://man7.org/linux/man-pages/man7/cpuset.7.html
func WithCPUSet(cpulist string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.CpusetCpus = cpulist
		return nil
	}
}

// WithMems configures the [set of memory nodes] on which processes of the
// container are allowed to allocate memory. The list of memory nodes is a
// comma-separated list of memory node numbers and ranges of numbers, where the
// numbers are decimal numbers. For instance, “1,3,5”, "0-42,666", et cetera.
//
// [set of memory nodes]: https://man7.org/linux/man-pages/man7/cpuset.7.html
func WithMems(memlist string) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		o.Opts.HostConfig.CpusetMems = memlist
		return nil
	}
}

// WithCustomInit instructs Docker to run an init inside the container that
// forwards signals and reaps processes.
func WithCustomInit() Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.HostConfig)
		customInit := true
		o.Opts.HostConfig.Init = &customInit
		return nil
	}
}

// WithUser configures the user and optionally group the command(s) inside a
// container will be run as, taking either a user name, user:group names, or a
// user ID. WithUser will remove any previously configured group, either setting
// the specified group or configuring no group at all (so that the container
// image default applies).
func WithUser[I identity.Principal](id I) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.User = identity.WithUser(id)
		return nil
	}
}

// WithGroup configures the group the command(s) inside a container will be run
// as, taking either a group name or ID. If an empty group name "" is specified,
// any configured group name or ID will be removed.
func WithGroup[I identity.Principal](gid I) Opt {
	return func(o *Options) error {
		ensure.Value(&o.Opts.Config)
		o.Opts.Config.User = identity.WithGroup(o.Opts.Config.User, gid)
		return nil
	}
}
