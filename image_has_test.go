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

	"github.com/docker/docker/api/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/morbyd/session"
	. "github.com/thediveo/success"
)

var _ = Describe("image presence", Ordered, func() {

	var sess *Session

	BeforeAll(func(ctx context.Context) {
		sess = Successful(NewSession(ctx,
			session.WithAutoCleaning("test.morbyd=")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
	})

	It("reports whether an image is locally available", func(ctx context.Context) {
		const imgref = "busybox:latestandgreatest"
		const imgreflatest = "busybox:latest"

		_, _ = sess.Client().ImageRemove(ctx, imgref, types.ImageRemoveOptions{})
		Expect(sess.HasImage(ctx, imgref)).To(BeFalse())

		Expect(sess.Client().ImagePull(ctx, imgreflatest, types.ImagePullOptions{})).Error().NotTo(HaveOccurred())
		Expect(sess.Client().ImageTag(ctx, imgreflatest, imgref))
		DeferCleanup(func(ctx context.Context) {
			Expect(sess.Client().ImageRemove(ctx, imgref, types.ImageRemoveOptions{})).Error().NotTo(HaveOccurred())
		})
		Expect(sess.HasImage(ctx, imgref)).To(BeTrue())
	})

	It("reports image listing errors", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		Expect(sess.HasImage(ctx, "busybox:absolutelygreatest")).Error().To(HaveOccurred())
	})

})
