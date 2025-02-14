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

package morbyd

import (
	"context"
	"fmt"
	"io"

	"github.com/containerd/platforms"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
)

// PushImageOpt is a configuration option to push a containerf image using
// [Session.PushImage].
type PushImageOpt func(*pushImageOptions) error

type pushImageOptions struct {
	image.PushOptions
	out io.Writer
}

// PushImage pushes a container image to a container registry.
func (s *Session) PushImage(ctx context.Context, image string, opts ...PushImageOpt) error {
	pios := pushImageOptions{}
	for _, opt := range opts {
		if err := opt(&pios); err != nil {
			return err
		}
	}
	if pios.out == nil {
		pios.out = io.Discard
	}
	r, err := s.moby.ImagePush(ctx, image, pios.PushOptions)
	if err != nil {
		return fmt.Errorf("image push failed, reason: %w", err)
	}
	defer r.Close()
	err = jsonmessage.DisplayJSONMessagesStream(r, pios.out, 0, false, nil)
	return err
}

// WithPushImageOutput specifies the writer to send the output of the image pull
// process to.
func WithPushImageOutput(w io.Writer) PushImageOpt {
	return func(pios *pushImageOptions) error {
		pios.out = w
		return nil
	}
}

// WithPushImageAllTags specifies that all tags of the image are to be pushed to
// the repository.
func WithPushImageAllTags() PushImageOpt {
	return func(pios *pushImageOptions) error {
		pios.All = true
		return nil
	}
}

// WithPushImagePlatform specifies to push a platform-specific manifest as a
// single-platform image to the registry.
func WithPushImagePlatform(platform string) PushImageOpt {
	return func(pios *pushImageOptions) error {
		p, err := platforms.Parse(platform)
		if err != nil {
			return err
		}
		pios.Platform = &p
		return nil
	}
}
