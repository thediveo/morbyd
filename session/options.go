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

package session

import (
	"fmt"
	"strings"

	"github.com/docker/docker/client"
	lbls "github.com/thediveo/morbyd/labels"
	"github.com/thediveo/morbyd/moby"
)

// Opt is a configuration option for creating sessions using
// [github.com/thediveo/morbyd.NewSession].
type Opt func(*Options) error

// Options stores the configuration options when creating a new testing session,
// including a Docker client.
type Options struct {
	Labels           map[string]string
	DockerClientOpts []client.Opt

	// If not "", then AutoCleaningLabel specifies a label that when stuck on
	// containers and networks identifies them for automatic cleaning before and
	// after test sessions. Please note that the auto-cleaning label is also
	// added to the session labels in order to automatically attach it to newly
	// created containers and networks.
	AutoCleaningLabel string

	// A function supplied by a test option to wrap the Docker client with
	// something else, such as a mock, double, or whatever you wanna call it.
	Wrapper func(moby.Client) moby.Client
}

// WithAutoCleaning enables autocleaning containers and networks before and
// after test sessions, in either “KEY=VALUE” or “KEY=” format. This label is
// automatically attached to any container and network created in a test
// session.
func WithAutoCleaning(label string) Opt {
	return func(o *Options) error {
		key, _, ok := strings.Cut(label, "=")
		if !ok || key == "" {
			return fmt.Errorf("auto cleaning label must be in format \"KEY=\" or \"KEY=VALUE\", "+
				"got %q", label)
		}
		o.AutoCleaningLabel = label
		ensureLabelsMap(o)
		return lbls.Labels(o.Labels).Add(label)
	}
}

// WithLabel specifies a single key-value label to be automatically attached to
// container images, containers, and networks created in this session. These
// labels can be used, for instance, to automatically clean up any left-over
// images, containers, and networks.
func WithLabel(label string) Opt {
	return func(o *Options) error {
		ensureLabelsMap(o)
		return lbls.Labels(o.Labels).Add(label)
	}
}

// WithLabels specifies multiple key-value labels to be automatically attached
// to container images, containers, and networks created in this session. These
// labels can be used, for instance, to automatically clean up any left-over
// images, containers, and networks.
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

// WithDockerOpts specifies additional options to apply when creating the Docker
// client as part of a new session.
func WithDockerOpts(opts ...client.Opt) Opt {
	return func(o *Options) error {
		o.DockerClientOpts = append(o.DockerClientOpts, opts...)
		return nil
	}
}
