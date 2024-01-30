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
	"time"
)

const DefaultSleep = 10 * time.Millisecond

// Sleep the specified duration, returning early in case the specified context
// gets cancelled. When cancelled, Sleep returns the context's error, otherwise
// when sleeping through the specified duration, it returns nil.
func Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		// We were cancelled during out nap, so let's call it a day.
		if !timer.Stop() {
			<-timer.C
		}
		return ctx.Err()
	case <-timer.C:
	}
	return nil
}
