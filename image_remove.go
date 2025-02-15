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

	"github.com/docker/docker/api/types/image"
)

// RemoveImageOpt is a configuration option to remove container image using
// [Session.RemoveImage].
type RemoveImageOpt func(*image.RemoveOptions)

// RemoveImage removes a container image.
func (s *Session) RemoveImage(ctx context.Context, imageid string, opts ...RemoveImageOpt) ([]image.DeleteResponse, error) {
	rios := image.RemoveOptions{}
	for _, opt := range opts {
		opt(&rios)
	}
	return s.moby.ImageRemove(ctx, imageid, rios)
}

func WithForceRemoveImage() RemoveImageOpt {
	return func(ro *image.RemoveOptions) {
		ro.Force = true
	}
}

func WithRemoveImagePruneChildren() RemoveImageOpt {
	return func(ro *image.RemoveOptions) {
		ro.PruneChildren = true
	}
}
