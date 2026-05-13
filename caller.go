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

package morbyd

import (
	"fmt"
	"runtime"
)

// caller returns a textual description of the call site, offset+1 calls down in
// the call stack; otherwise, an empty string.
func caller(stackskip uint, lineoffset int) string {
	pc, file, line, ok := runtime.Caller(1 + int(stackskip))
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return fmt.Sprintf("%s:%d", file, line+lineoffset)
	}
	return fmt.Sprintf("%s:%d (%s)", file, line+lineoffset, fn.Name())
}
