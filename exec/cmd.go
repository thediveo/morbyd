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

package exec

type Cmd []string

// Command takes a command with its optional arguments, returning a Cmd object
// to be used with [github.com/thediveo/morbyd.Container.Exec].
func Command(cmd string, args ...string) Cmd {
	return append(Cmd{cmd}, args...)
}
