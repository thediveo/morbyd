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

package build

import (
	"io"

	"github.com/docker/docker/api/types"
	lbls "github.com/thediveo/morbyd/labels"
)

// Opt is a configuration option to build a container image using
// [github.com/thediveo/morbyd.Session.BuildImage].
type Opt func(*Options) error

// Options represents the plethora of configuration options when building a
// container image, as well as additional configuration options for handling the
// output of build processes.
//
// See also: [Build an image (Docker API)].
//
// [Build an image (Docker API)]: https://docs.docker.com/engine/api/v1.43/#tag/Image/operation/ImageBuild
type Options struct {
	Out io.Writer
	types.ImageBuildOptions
}

// WithTag specifies a name and optionally tag in “name:tag” format. This option
// can be specified multiple times so that the built image will be tagged with
// these multiple tags. If the “:tag” part is omitted, the default “:latest” is
// assumed.
func WithTag(nametag string) Opt {
	return func(o *Options) error {
		o.Tags = append(o.Tags, nametag)
		return nil
	}
}

// WithBuildArg specifies an image build argument, in either of the three forms
// “foo=bar”, “foo=” and “foo”. Please note that “foo=” and “foo” specify
// different things: “foo=” specifies a build arg named “foo” with an empty
// value “”, whereas “foo” specifies an unset (null valued) build arg named
// “foo”. WithBuildArg can be specified multiple times.
func WithBuildArg(barg string) Opt {
	return func(o *Options) error {
		if o.BuildArgs == nil {
			o.BuildArgs = map[string]*string{}
		}
		return BuildArgs(o.BuildArgs).Add(barg)
	}
}

// WithBuildArgs specifies multiple image build arguments. See also:
// [WithBuildArg].
func WithBuildArgs(bargs ...string) Opt {
	return func(o *Options) error {
		if o.BuildArgs == nil {
			o.BuildArgs = map[string]*string{}
		}
		for _, barg := range bargs {
			if err := BuildArgs(o.BuildArgs).Add(barg); err != nil {
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

// WithLabel adds a key-value label to the built image.
func WithLabel(label string) Opt {
	return func(o *Options) error {
		ensureLabelsMap(o)
		return lbls.Labels(o.Labels).Add(label)
	}
}

// WithLabels adds multiple key-value labels to the built image.
func WithLabels(labels ...string) Opt {
	return func(o *Options) error {
		ensureLabelsMap(o)
		for _, label := range labels {
			if err := lbls.Labels(o.Labels).Add(label); err != nil {
				return err
			}
		}
		return nil
	}
}

// ensureLabelsMap is a helper to ensure that the Options.Labels map is
// initialized.
func ensureLabelsMap(o *Options) {
	if o.Labels != nil {
		return
	}
	o.Labels = map[string]string{}
}

// WithoutCache specifies to never use the cache when building the image.
func WithoutCache() Opt {
	return func(o *Options) error {
		o.NoCache = true
		return nil
	}
}

// WithPullAlways attempts to pull the parent image even if an older image
// exists locally.
func WithPullAlways() Opt {
	return func(o *Options) error {
		o.PullParent = true
		return nil
	}
}

// WithSquash creates a new image from the parent with all changes applied to
// this single new layer.
func WithSquash() Opt {
	return func(o *Options) error {
		o.Squash = true
		return nil
	}
}

// WithRemoveIntermediateContainers removes intermediate containers after a
// successful build.
func WithRemoveIntermediateContainers() Opt {
	return func(o *Options) error {
		o.Remove = true
		return nil
	}
}

// WithAlwaysRemoveIntermediateContainers always removes intermediate
// containers, even upon image build failure.
func WithAlwaysRemoveIntermediateContainers() Opt {
	return func(o *Options) error {
		o.ForceRemove = true
		return nil
	}
}

// WithOutput set the writer to which the output of the image build process is
// sent to.
func WithOutput(w io.Writer) Opt {
	return func(o *Options) error {
		o.Out = w
		return nil
	}
}

// WithOpts (re)sets all image build options to the specified settings at once,
// with the sole exception of an optional output writer for image build process
// output (that needs to set independently using WithImageBuildOutput).
func WithOpts(opts types.ImageBuildOptions) Opt {
	return func(o *Options) error {
		o.ImageBuildOptions = opts
		return nil
	}
}
