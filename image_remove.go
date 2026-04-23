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

	"github.com/moby/moby/client"

	"github.com/thediveo/morbyd/v2/remove"
)

// RemoveImage removes a container image.
func (s *Session) RemoveImage(ctx context.Context, imageid string, opts ...remove.Opt) (client.ImageRemoveResult, error) {
	rios := remove.Options{}
	for _, opt := range opts {
		if err := opt(&rios); err != nil {
			return client.ImageRemoveResult{}, err
		}
	}
	return s.moby.ImageRemove(ctx, imageid, rios.ImageRemoveOptions)
}
