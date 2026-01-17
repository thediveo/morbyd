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

	"github.com/containerd/errdefs"
)

// HasImage returns true if the image referenced by imageref is locally
// available, otherwise false.
func (s *Session) HasImage(ctx context.Context, imageref string) (bool, error) {
	_, err := s.moby.ImageInspect(ctx, imageref)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("image inspection failed, reason: %w", err)
	}
	return true, nil
}
