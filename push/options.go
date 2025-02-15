// Copyright 2025 Harald Albrecht.
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

package push

import (
	"io"

	"github.com/containerd/platforms"
	"github.com/docker/docker/api/types/image"
)

// Opt is a configuration option to push a container image using
// [github.com/thediveo/morbyd.Session.PushImage].
type Opt func(*Options) error

// Options represent the configuration options when pushing a container image,
// as well as additional configuration options for handling the output of push
// processes.
type Options struct {
	Out io.Writer
	image.PushOptions
}

// WithOutput specifies the writer to send the output of the image push process
// to.
func WithOutput(w io.Writer) Opt {
	return func(o *Options) error {
		o.Out = w
		return nil
	}
}

// WithAllTags specifies that all tags of the image are to be pushed to the
// repository.
func WithAllTags() Opt {
	return func(o *Options) error {
		o.All = true
		return nil
	}
}

// WithPlatform specifies to push a platform-specific manifest as a
// single-platform image to the registry.
func WithPlatform(platform string) Opt {
	return func(o *Options) error {
		p, err := platforms.Parse(platform)
		if err != nil {
			return err
		}
		o.Platform = &p
		return nil
	}
}

// WithRegistryAuth specifies the base64 encoded credentials for the registry to
// push the image to.
func WithRegistryAuth(base64cred string) Opt {
	return func(o *Options) error {
		o.RegistryAuth = base64cred
		return nil
	}
}
