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

package exec

import (
	"io"

	"github.com/docker/docker/api/types"
)

// Opt is a configuration option to run a command inside a container using
// [github.com/thediveo/morbyd.Container.Run].
type Opt func(*Options) error

// Options represents the plethora of configuration options when running a
// command inside a container, as well as additional configuration options for
// handling the output (and optionally input) stream of the command.
//
// Please note that the command to be executed is not passed as an option to
// [github.com/thediveo/morbyd.Container.Exec], but instead as a required [Cmd].
type Options struct {
	In   io.Reader
	Out  io.Writer
	Err  io.Writer
	Conf types.ExecConfig
}

// WithCombinedOutput sends the commands's stdout and stderr to the specified
// io.Writer. This also automatically attaches the commands's stdout and stderr
// after the command execution context has been created.
func WithCombinedOutput(w io.Writer) Opt {
	return func(o *Options) error {
		o.Out = w
		o.Err = w
		return nil
	}
}

// WithDemuxedOutput disables the TTY option in order to send the command's
// stdout and stderr properly separated to the specified out and err io.Writer.
//
// It is a known limitation of the (pseudo) TTY to combine both stdout and
// stderr of the commands's output streams, and there is no way for Docker (and
// thus for us) to demultiplex this output sludge after the fact.
func WithDemuxedOutput(out io.Writer, err io.Writer) Opt {
	return func(o *Options) error {
		o.Conf.Tty = false
		o.Out = out
		o.Err = err
		return nil
	}
}

// WithInput sends input data from the specified io.Reader to the commands's
// stdin. For this, it allocates a stdin for the command when creating its
// execution context and then attaches to this stdin after creation.
func WithInput(r io.Reader) Opt {
	return func(o *Options) error {
		o.In = r
		return nil
	}
}

// WithTTY allocates a pseudo TTY for the commands's input and output.
//
// Please note that using a TTY causes the commands's stdout and stderr streams
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

// WithEnvVars adds multiple environment variables to the container to be
// started.
func WithEnvVars(vars ...string) Opt {
	return func(o *Options) error {
		o.Conf.Env = append(o.Conf.Env, vars...)
		return nil
	}
}

// WithWorkingDir sets the (current) working directory when executing the
// command inside the container.
func WithWorkingDir(path string) Opt {
	return func(o *Options) error {
		o.Conf.WorkingDir = path
		return nil
	}
}

// WithUser sets the user (ID or name) that executes the command.
func WithUser(user string) Opt {
	return func(o *Options) error {
		o.Conf.User = user
		return nil
	}
}

// WithPrivileged executes the command as privileged inside the container.
func WithPrivileged() Opt {
	return func(o *Options) error {
		o.Conf.Privileged = true
		return nil
	}
}

// WithConsoleSize sets with width and height of the pseudo TTY; please note the
// width-height order in contrast to the Docker API order.
func WithConsoleSize(width, height uint) Opt {
	return func(o *Options) error {
		o.Conf.ConsoleSize = &[2]uint{height, width} // sic!
		return nil
	}
}
