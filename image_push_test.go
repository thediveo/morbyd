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
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"time"

	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/success"
)

var _ = Describe("pushing images", Ordered, func() {

	BeforeEach(func(ctx context.Context) {
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
		})
	})

	It("pushes an image", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ImagePush")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rc := io.NopCloser(strings.NewReader(`
{"status":"foobar"}
`))
		rec.ImagePush(Any, Any, Any).Return(rc, nil)

		var buff bytes.Buffer
		Expect(sess.PushImage(ctx, "buzzybocks:earliest", WithPushImageOutput(&buff))).To(Succeed())
		Expect(buff.String()).To(Equal("foobar\n"))
	})

	It("reports API errors", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ImagePush")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()
		rec.ImagePush(Any, Any, Any).Return(nil, errors.New("error IJK305I"))

		Expect(sess.PushImage(ctx, "buzzybocks:earliest")).NotTo(Succeed())
	})

	It("reports stream errors", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ImagePush")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()
		rc := io.NopCloser(strings.NewReader(`
{"errorDetail":{"code":666,"message":"error IJK305I"}}
`))
		rec.ImagePush(Any, Any, Any).Return(rc, nil)

		Expect(sess.PushImage(ctx, "buzzybocks:earliest")).NotTo(Succeed())
	})

	It("reports options errors before attempting to push", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ImagePush")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()
		rec.ImagePush(Any, Any, Any).Times(0)

		Expect(sess.PushImage(ctx, "buzzybocks:earliest",
			WithPushImagePlatform("arm-selig"))).Error().To(
			MatchError(ContainSubstring(`"arm-selig": unknown operating system or architecture`)))
	})

	Context("options", func() {

		It("pushes all tags", func() {
			var pios pushImageOptions
			Expect(WithPushImageAllTags()(&pios)).To(Succeed())
			Expect(pios.All).To(BeTrue())
		})

		It("specifies a platform", func() {
			var pios pushImageOptions
			Expect(WithPushImagePlatform("leinucks/arm64")(&pios)).To(Succeed())
			Expect(pios.Platform).To(And(
				HaveField("OS", "leinucks"),
				HaveField("Architecture", "arm64"),
			))
		})

	})

})
