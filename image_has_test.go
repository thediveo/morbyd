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
	"errors"
	"time"

	image "github.com/docker/docker/api/types/image"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/success"
)

var _ = Describe("image presence", Ordered, func() {

	Context("image presence checking", Ordered, func() {

		BeforeEach(func(ctx context.Context) {
			goodgos := Goroutines()
			DeferCleanup(func() {
				Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
			})
		})

		It("reports whether an image is locally available", func(ctx context.Context) {
			const imgref = "busybox:latestandgreatest"

			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ImageList")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()
			rec.ImageList(Any, Any).Return([]image.Summary{}, nil)
			rec.ImageList(Any, Any).Return([]image.Summary{
				{ /*doesn't matter what, just needs to exist*/ },
			}, nil)

			Expect(sess.HasImage(ctx, imgref)).To(BeFalse())
			Expect(sess.HasImage(ctx, imgref)).To(BeTrue())
		})

		It("reports image listing errors", func(ctx context.Context) {
			ctrl := mock.NewController(GinkgoT())
			sess := Successful(NewSession(ctx,
				WithMockController(ctrl, "ImageList")))
			DeferCleanup(func(ctx context.Context) {
				sess.Close(ctx)
			})
			rec := sess.Client().(*MockClient).EXPECT()
			rec.ImageList(Any, Any).Return(nil, errors.New("error IJK305I"))

			Expect(sess.HasImage(ctx, "busybox:absolutelygreatest")).Error().To(HaveOccurred())
		})

	})

})
