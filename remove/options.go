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

package remove

import (
	"github.com/containerd/platforms"
	"github.com/moby/moby/client"
)

// Opt is a configuration option to remove a container image using
// [github.com/thediveo/morbyd/v2.Session.Run]. Please see also [Options] for
// more information.
type Opt func(*Options) error

// Options represents the configuration options when removing a container image.
type Options struct {
	client.ImageRemoveOptions
}

// WithPlatform configures to remove an image only for the specified platform(s)
// variants; WithPlatform can be used multiple times, adding more platforms
// variants as necessary.
func WithPlatform(platform string) Opt {
	return func(o *Options) error {
		pp, err := platforms.Parse(platform)
		if err != nil {
			return err
		}
		o.Platforms = append(o.Platforms, pp)
		return nil
	}
}

// WithForce forces image removal.
func WithForce() Opt {
	return func(o *Options) error {
		o.Force = true
		return nil
	}
}

// WithSchwartz forces The Force; see also [WithForce].
func WithSchwartz() Opt { return WithForce() }

// WithPruneChildren configures removing untagged parent images. Yes, that's
// what the [docker image rm] docs say.
//
// [docker image rm]: https://docs.docker.com/reference/cli/docker/image/rm/#options
func WithPruneChildren() Opt {
	return func(o *Options) error {
		o.PruneChildren = true
		return nil
	}
}
