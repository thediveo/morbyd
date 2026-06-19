// Copyright 2026 Harald Albrecht.
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

package imgrefs

import (
	"io"

	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"

	"github.com/thediveo/morbyd/v2/build"
	"github.com/thediveo/morbyd/v2/internal/ensure"
)

// Opt is a configuration option to determine the Docker images referenced by a
// Dockerfile.
type Opt func(*Options) error

// Options represents the configuration options for analyzing Dockerfiles using
// buildkit.
type Options struct {
	BuildArgs  build.BuildArgs
	Dockerfile string
	Out        io.Writer
	dockerfile2llb.ConvertOpt
}

// WithBuildArg specifies an image build argument, in either of the three forms
// “foo=bar”, “foo=” and “foo”. Please note that “foo=” and “foo” specify
// different things: “foo=” specifies a build arg named “foo” with an empty
// value “”, whereas “foo” specifies an unset (null valued) build arg named
// “foo”. WithBuildArg can be specified multiple times.
func WithBuildArg(barg string) Opt {
	return func(o *Options) error {
		ensure.Map(&o.BuildArgs)
		return o.BuildArgs.Add(barg)
	}
}

// WithBuildArgs specifies multiple image build arguments. See also:
// [WithBuildArg].
func WithBuildArgs(bargs ...string) Opt {
	return func(o *Options) error {
		ensure.Map(&o.BuildArgs)
		for _, barg := range bargs {
			if err := o.BuildArgs.Add(barg); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithDockerfile specifies the name of the Dockerfile.
func WithDockerfile(name string) Opt {
	return func(o *Options) error {
		o.Dockerfile = name
		return nil
	}
}

// WithOutput set the writer to which the output of the Dockerfile analysis
// process is sent to.
func WithOutput(w io.Writer) Opt {
	return func(o *Options) error {
		o.Out = w
		return nil
	}
}
