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

package ensure

// Map creates the map to which m points to, but only if it hasn't been created
// already.
func Map[K comparable, V any](m *map[K]V) {
	if *m != nil {
		return
	}
	*m = make(map[K]V)
}

// Value creates the zero value that v indirectly points to, but only if it
// hasn't been created already.
func Value[T any, PPT interface{ **T }](v PPT) {
	if *v != nil {
		return
	}
	*v = new(T)
}
