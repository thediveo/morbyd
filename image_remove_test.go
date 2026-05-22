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
	"context"
	"errors"

	"github.com/thediveo/morbyd/v2/remove"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("image removal", Ordered, func() {

	It("reports option errors", func() {
		res, err := (&Session{}).RemoveImage(
			context.Background(), "",
			func(o *remove.Options) error { return errors.New("JKL305") })
		Expect(err).To(HaveOccurred())
		Expect(res).To(BeZero())
	})

})
