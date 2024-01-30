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
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	lbls "github.com/thediveo/morbyd/labels"
	"github.com/thediveo/morbyd/run/dockercli"
	"github.com/thediveo/morbyd/strukt"
	"golang.org/x/exp/maps"
)

// Opt is a configuration option to run a container using
// [github.com/thediveo/morbyd.Session.Run]. Please see also [Options] for
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
	Name string
	In   io.Reader
	Out  io.Writer
	Err  io.Writer
	Conf container.Config
	Host container.HostConfig
	Net  network.NetworkingConfig
}

// WithCombinedOutput sends the container's stdout and stderr to the specified
// io.Writer. This also automatically attaches the container's stdout and stderr
// after the container has been created.
func WithCombinedOutput(w io.Writer) Opt {
	return func(o *Options) error {
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
		o.Conf.Tty = false
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
		o.Conf.OpenStdin = true
		o.Conf.StdinOnce = true
		o.In = r
		return nil
	}
}

// WithName sets the optional name of the container to create.
func WithName(name string) Opt {
	return func(o *Options) error {
		o.Name = name
		return nil
	}
}

// WithCommand sets the optional command to execute at container start.
func WithCommand(cmd ...string) Opt {
	return func(o *Options) error {
		o.Conf.Cmd = strslice.StrSlice(cmd)
		return nil
	}
}

// WithEnvVars adds multiple environment variables to the container to be
// started.
func WithEnvVars(vars ...string) Opt {
	return func(o *Options) error {
		o.Conf.Env = append(o.Conf.Env, vars...)
		return nil
	}
}

// ClearLabels clears any labels inherited from the [Session] when creating a
// new container.
func ClearLabels() Opt {
	return func(o *Options) error {
		maps.Clear(o.Conf.Labels)
		return nil
	}
}

// WithLabel adds a label in “key=value” format to the container's labels.
func WithLabel(label string) Opt {
	return func(o *Options) error {
		ensureLabelsMap(o)
		return lbls.Labels(o.Conf.Labels).Add(label)
	}
}

// WithLabels adds multiple labels in “key=value” format to the container's
// labels.
func WithLabels(labels ...string) Opt {
	return func(o *Options) error {
		ensureLabelsMap(o)
		for _, label := range labels {
			if err := lbls.Labels(o.Conf.Labels).Add(label); err != nil {
				return err
			}
		}
		return nil
	}
}

func ensureLabelsMap(o *Options) {
	if o.Conf.Labels != nil {
		return
	}
	o.Conf.Labels = map[string]string{}
}

// WithStopSignal sets the name of the signal to be sent to its initial process
// when stopping the container.
func WithStopSignal(s string) Opt {
	return func(o *Options) error {
		o.Conf.StopSignal = s
		return nil
	}
}

// WithStopTimeout sets the timeout (in seconds) to stop the container.
func WithStopTimeout(secs int) Opt {
	return func(o *Options) error {
		o.Conf.StopTimeout = &secs
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
		o.Conf.Tty = true
		return nil
	}
}

// WithAutoRemove removes the container after it has stopped.
func WithAutoRemove() Opt {
	return func(o *Options) error {
		o.Host.AutoRemove = true
		return nil
	}
}

// WithPrivileged runs the container as privileged, including all capabilities
// and not masking out certain filesystem elements.
func WithPrivileged() Opt {
	return func(o *Options) error {
		o.Host.Privileged = true
		return nil
	}
}

// WithCapAdd adds an individual kernel capabilities to the initial process in
// the container.
func WithCapAdd(cap string) Opt {
	return func(o *Options) error {
		o.Host.CapAdd = append(o.Host.CapAdd, cap)
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
		o.Host.CapDrop = strslice.StrSlice{"ALL"}
		return nil
	}
}

// WithCgroupnsMode configures the cgroup namespace to use when creating the new
// container; it can be either “” (empty, use the daemon's default)
// configuration, “private”, or “host”.
func WithCgroupnsMode(mode string) Opt {
	return func(o *Options) error {
		o.Host.CgroupnsMode = container.CgroupnsMode(mode)
		return nil
	}
}

// WithIPCMode configures the IPC namespace to use when creating the new
// container; it can be either “” (empty, use the daemon's default), “none”
// (private, but without /dev/shm mounted), “private”, “shareable”, “host”, or
// container:NAMEID”.
func WithIPCMode(mode string) Opt {
	return func(o *Options) error {
		o.Host.IpcMode = container.IpcMode(mode)
		return nil
	}
}

// WithNetworkMode configures the net(work) namespace to use when creating the new
// container; it can be either “” (create a new one), “none”, “host”, or
// container:NAMEID”.
func WithNetworkMode(mode string) Opt {
	return func(o *Options) error {
		o.Host.NetworkMode = container.NetworkMode(mode)
		return nil
	}
}

// WithPIDMode configures the PID namespace to use when creating the new
// container; it can be either “” (create a new one), “host”, or
// container:NAMEID”.
func WithPIDMode(mode string) Opt {
	return func(o *Options) error {
		o.Host.PidMode = container.PidMode(mode)
		return nil
	}
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
			if o.Conf.Volumes == nil {
				o.Conf.Volumes = map[string]struct{}{}
			}
			o.Conf.Volumes[vol] = struct{}{}
			return nil
		}
	}

	// All other volume specs are now handled via the host configuration bind
	// (mounts).
	if parsedVol.Type == string(mount.TypeBind) {
		vol = bindVolumeToBind(vol)
	}
	return func(o *Options) error {
		o.Host.Binds = append(o.Host.Binds, vol)
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
		o.Host.Mounts = append(o.Host.Mounts, mount)
		return nil
	}
}

// WithTmpfs specifies the path inside the container to mount a new tmpfs
// instance on, using default options (unlimited size, world-writable).
func WithTmpfs(path string) Opt {
	return func(o *Options) error {
		if o.Host.Tmpfs == nil {
			o.Host.Tmpfs = map[string]string{}
		}
		o.Host.Tmpfs[path] = ""
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
		if o.Host.Tmpfs == nil {
			o.Host.Tmpfs = map[string]string{}
		}
		o.Host.Tmpfs[path] = opts
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
		o.Host.Devices = append(o.Host.Devices, container.DeviceMapping{
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
		o.Host.ReadonlyRootfs = true
		return nil
	}
}

// WithSecurityOpt configures an additional security option, such as
// “seccomp=unconfined”.
func WithSecurityOpt(opt string) Opt {
	return func(o *Options) error {
		o.Host.SecurityOpt = append(o.Host.SecurityOpt, opt)
		return nil
	}
}

// WithConsoleSize sets with width and height of the pseudo TTY; please note the
// width-height order in contrast to the Docker API order.
func WithConsoleSize(width, height uint) Opt {
	return func(o *Options) error {
		o.Host.ConsoleSize = [2]uint{height, width} // sic!
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
func WithNetwork(net string) Opt {
	return func(o *Options) error {
		dry := dockercli.NetworkOpt{}
		if err := dry.Set(net); err != nil {
			return fmt.Errorf("malformed WithNetwork parameter %q, reason: %s",
				net, err)
		}
		ep := dry.Value()[0]
		if o.Net.EndpointsConfig == nil {
			o.Net.EndpointsConfig = map[string]*network.EndpointSettings{}
		}
		o.Net.EndpointsConfig[ep.Target] = &network.EndpointSettings{
			NetworkID:         ep.Target,
			Aliases:           ep.Aliases,
			DriverOpts:        ep.DriverOpts,
			Links:             ep.Links,
			IPAddress:         ep.IPv4Address,
			GlobalIPv6Address: ep.IPv6Address, // TODO: clarify w. respect to GlobalIPv6PrefixLen
			MacAddress:        ep.MacAddress,
		}
		return nil
	}
}

// WithHostname configures the host name to use inside the container.
func WithHostname(host string) Opt {
	return func(o *Options) error {
		o.Conf.Hostname = host
		return nil
	}
}

// WithRestartPolicy configures the restart policy (“no”, “always”,
// “on-failure”, “unless-stopped”) as well as the maximum attempts at restarting
// the container.
func WithRestartPolicy(policy string, maxretry int) Opt {
	return func(o *Options) error {
		o.Host.RestartPolicy = container.RestartPolicy{
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
		o.Host.PublishAllPorts = true
		return nil
	}
}

func WithCustomInit() Opt {
	return func(o *Options) error {
		customInit := true
		o.Host.Init = &customInit
		return nil
	}
}
