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

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/thediveo/morbyd/push"
)

// PushImage pushes a container image to a container registry.
func (s *Session) PushImage(ctx context.Context, image string, opts ...push.Opt) error {
	popts := push.Options{}
	for _, opt := range opts {
		if err := opt(&popts); err != nil {
			return err
		}
	}
	if popts.Out == nil {
		popts.Out = io.Discard
	}
	r, err := s.moby.ImagePush(ctx, image, popts.PushOptions)
	if err != nil {
		return fmt.Errorf("image push failed, reason: %w", err)
	}
	defer r.Close()
	err = jsonmessage.DisplayJSONMessagesStream(r, popts.Out, 0, false, nil)
	return err
}
