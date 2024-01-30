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

package strukt

import (
	"fmt"
	"reflect"
	"strings"
)

// Unmarshal a string with multiple delimited fields into the given struct,
// returning nil on success. The fields of the struct are filled in their
// sequence of definition.
//
// If there are less fields in the string than struct fields, the remaining
// struct fields will be left as is. Unmarshal returns an error if there are
// more delemited fields than struct fields.
func Unmarshal(s string, delim string, strukt any) error {
	typ := reflect.TypeOf(strukt)
	if typ.Kind() != reflect.Pointer {
		return fmt.Errorf("strukt.Unmarshal: expected *struct { ... }, got %T", strukt)
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("strukt.Unmarshal: expected *struct { ... }, got %T", strukt)
	}
	fields := strings.Split(s, delim)
	if len(fields) > typ.NumField() {
		return fmt.Errorf("strukt.Unmarshal: too many fields for struct type %T", strukt)
	}
	structval := reflect.ValueOf(strukt).Elem()
	for idx, field := range fields {
		fieldval := structval.Field(idx)
		if !fieldval.CanSet() {
			return fmt.Errorf("strukt.Unmarshal: cannot set field %s", structval.Type().Field(idx).Name)
		}
		if fieldval.Type().Kind() != reflect.String {
			return fmt.Errorf("strukt.Unmarshal: expected field %s to have type string, got %s",
				structval.Type().Field(idx).Name, fieldval.Type().Name())
		}
		fieldval.SetString(field)
	}
	return nil
}
