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
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/thediveo/morbyd/build"
	"github.com/thediveo/morbyd/session"
	mock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("build image", Ordered, func() {

	Context(".dockerignore", Ordered, func() {

		BeforeEach(func() {
			goodfds := Filedescriptors()
			DeferCleanup(func() {
				Eventually(Filedescriptors).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
					ShouldNot(HaveLeakedFds(goodfds))
			})
		})

		It("returns an empty ignore list when file is missing", func() {
			Expect(readIgnorePatterns("_test/dockerignore/notexisting")).To(BeZero())
		})

		It("returns the correct ignore list", func() {
			Expect(readIgnorePatterns("_test/dockerignore/.dockerignore")).To(ConsistOf(
				"foo",
				// note: leading forward slashes are removed from Docker ignore
				// patterns; see also:
				// https://github.com/moby/patternmatcher/blob/347bb8d8d557f90d1b75cd8bca3c0177f380a979/ignorefile/ignorefile.go#L22
				"bar",
				"baz*",
			))
		})

		It("reports error when trying to read anon-existing ignore pattern file", func() {
			Expect(readIgnorePatterns("/dev/zero")).To(BeZero())
		})

	})

	When("building container images", Ordered, func() {

		var sess *Session

		BeforeAll(func(ctx context.Context) {
			sess = Successful(NewSession(ctx,
				session.WithAutoCleaning("test.morbyd=")))
			DeferCleanup(func(ctx context.Context) {
				// not strictly necessary as we're doing it anyway after each
				// individual test in order to check for leaked go routines.
				sess.Close(ctx)
			})
		})

		BeforeEach(func(ctx context.Context) {
			goodgos := Goroutines()
			goodfds := Filedescriptors()
			DeferCleanup(func() {
				sess.Close(ctx)
				Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
				Eventually(Filedescriptors).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
					ShouldNot(HaveLeakedFds(goodfds))
			})
		})

		It("rejects invalid options", func(ctx context.Context) {
			Expect(sess.BuildImage(ctx, "foobar",
				build.WithLabel(""))).Error().To(HaveOccurred())
		})

		It("builds an image and finds the correct stage build output canary", func(ctx context.Context) {
			const imageref = "morbyd/buzzybocks"

			_, _ = sess.Client().ImageRemove(ctx, imageref, types.ImageRemoveOptions{})

			var buff bytes.Buffer
			id := Successful(sess.BuildImage(ctx, "_test/buzzybocks",
				build.WithTag(imageref),
				build.WithBuildArg("HELLO=WORLD"),
				build.WithOutput(io.MultiWriter(GinkgoWriter, &buff)),
			))
			Expect(id).NotTo(BeEmpty())
			Expect(sess.Client().ImageRemove(
				ctx, imageref, types.ImageRemoveOptions{})).Error().To(
				Succeed())
			Expect(buff.String()).To(ContainSubstring("..WORLD.."))
		})

		It("fails to build an image with a failing Dockerfile", func(ctx context.Context) {
			const imageref = "morbyd/broken"

			id, err := sess.BuildImage(ctx, "_test/broken",
				build.WithTag(imageref),
			)
			Expect(err).To(HaveOccurred())
			Expect(id).To(BeEmpty())
			Expect(sess.Client().ImageRemove(
				ctx, imageref, types.ImageRemoveOptions{})).Error().To(
				MatchError(ContainSubstring("from daemon: No such image: " + imageref)))
		})

		It("fails to build an image with a non-existing Dockerfile", func(ctx context.Context) {
			const imageref = "morbyd/broken"

			id, err := sess.BuildImage(ctx, "_test/broken",
				build.WithDockerfile("Mobyfile"),
				build.WithTag(imageref),
			)
			Expect(err).To(HaveOccurred())
			Expect(id).To(BeEmpty())

		})

		It("reports when the build context cannot be built", func(ctx context.Context) {
			const imageref = "morbyd/broken"

			id, err := sess.BuildImage(ctx, "_test/broken-dockerignore",
				build.WithTag(imageref),
			)
			Expect(err).To(HaveOccurred())
			Expect(id).To(BeEmpty())
		})

	})

	It("skips invalid aux messages", func(ctx context.Context) {
		ctrl := mock.NewController(GinkgoT())
		sess := Successful(NewSession(ctx,
			WithMockController(ctrl, "ImageBuild")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		rec := sess.Client().(*MockClient).EXPECT()

		rc := io.NopCloser(strings.NewReader(`
{"aux":{"ID":"foobar"}}
{"aux":{"ID":""}}
{"aux":{"ID":42}}
`))
		rec.ImageBuild(Any, Any, Any).Return(types.ImageBuildResponse{Body: rc}, nil)

		Expect(sess.BuildImage(ctx, "./_test/dockerignore")).To(Equal("foobar"))
	})

})
