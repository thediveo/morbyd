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

package morbyd

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/thediveo/morbyd/pull"
)

// PullImage pulls a container image specified by the image reference, if not
// already locally available. The additional pull options are applied in the
// order they are provided.
//
// If no pull process output writer has been specified using [pull.WithOutput]
// any output (such as pull progress, et cetera) will simply be discarded.
//
// Any pull process errors will be reported.
func (s *Session) PullImage(ctx context.Context, imgref string, opts ...pull.Opt) error {
	pios := pull.Options{}
	for _, opt := range opts {
		if err := opt(&pios); err != nil {
			return err
		}
	}
	if pios.Out == nil {
		pios.Out = io.Discard
	}
	r, err := s.moby.ImagePull(ctx, imgref, pios.PullOptions)
	if err != nil {
		return fmt.Errorf("image pull failed, reason: %w", err)
	}
	defer r.Close() //nolint:errcheck // any error is irrelevant at this point
	err = jsonmessage.DisplayJSONMessagesStream(r, pios.Out, 0, false, nil)
	return err
}
