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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/registry"
	"github.com/thediveo/morbyd/pull"
	"github.com/thediveo/morbyd/push"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"

	"golang.org/x/crypto/bcrypt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/success"
)

const (
	registryPort  = 5999                      // port where we expose our local registry on loopback.
	originalImage = "busybox:latest"          // the upstream original image to get from the Docker registry.
	canaryImage   = "morbyd-busybox:weirdest" // the tag we'll use in the tests

	user     = "silly-user"
	password = "silly-password"
)

// full image reference to our local registry-located testing image
var localCanaryImage = fmt.Sprintf("127.0.0.1:%d/%s", registryPort, canaryImage)

var _ = Describe("given a (local) registry", Ordered, Serial, func() {

	var magicauth string

	BeforeAll(func(ctx context.Context) {
		sess := Successful(NewSession(ctx,
			session.WithAutoCleaning("test=morbyd.registry")))
		DeferCleanup(func(ctx context.Context) {
			By("removing the local container registry")
			sess.Close(ctx)
		})

		By("creating a silly htpasswd")
		hash := Successful(bcrypt.GenerateFromPassword([]byte(password), 10))
		tmpDir := Successful(os.MkdirTemp("", "registry-silly-*"))
		DeferCleanup(os.RemoveAll, tmpDir)
		htpasswdPath := filepath.Join(tmpDir, "htpasswd")
		os.WriteFile(htpasswdPath,
			fmt.Appendf(nil, "%s:%s\n", user, hash),
			0644)

		By("starting a local container registry")
		Expect(sess.Run(ctx, "registry:3",
			run.WithName("local-registry"),
			run.WithPublishedPort(fmt.Sprintf("127.0.0.1:%d:5000", registryPort)),
			run.WithVolume(htpasswdPath+":/auth/htpasswd:ro"),
			run.WithEnvVars("REGISTRY_LOG_LEVEL=info",
				"REGISTRY_AUTH=htpasswd",
				"REGISTRY_AUTH_HTPASSWD_REALM=local-registry",
				"REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd",
				"REGISTRY_HTTP_SECRET=not-secret"),
			run.WithAutoRemove(),
			run.WithCombinedOutput(GinkgoWriter),
		)).Error().NotTo(HaveOccurred())

		By("waiting for the registry to become operational")
		Eventually(func() error {
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/v2/", registryPort))
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
				return fmt.Errorf("HTTP status code %d", resp.StatusCode)
			}
			return nil
		}).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
			Should(Succeed())

		magicauth = base64.URLEncoding.EncodeToString(
			Successful(json.Marshal(registry.AuthConfig{
				Username: user,
				Password: password,
			})))

	})

	BeforeEach(func(ctx context.Context) {
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
		})
	})

	It("pushes an image, needing auth", func(ctx context.Context) {
		sess := Successful(NewSession(ctx))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("pulling the canary image into the local Docker, if not already available")
		// normal PullImage will always first check instead of skipping
		// immediately, so we need to check explicitly before pulling.
		if !Successful(sess.HasImage(ctx, originalImage)) {
			Expect(sess.PullImage(ctx,
				originalImage,
				pull.WithRegistryAuth(magicauth),
				pull.WithOutput(timestamper.New(GinkgoWriter)))).To(Succeed())
		}
		By("tagging the canary image for local registry")
		Expect(sess.TagImage(ctx, originalImage, localCanaryImage)).To(Succeed())

		By("pushing the canary image into the local registry, without auth")
		Expect(sess.PushImage(ctx, localCanaryImage,
			push.WithOutput(timestamper.New(GinkgoWriter)))).NotTo(Succeed())

		By("pushing the canary image into the local registry, with silly auth")
		Expect(sess.PushImage(ctx, localCanaryImage,
			push.WithRegistryAuth(magicauth),
			push.WithOutput(timestamper.New(GinkgoWriter)))).To(Succeed())
	})

	It("pulls an image without auth", func(ctx context.Context) {
		sess := Successful(NewSession(ctx))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		By("ensuring the image isn't available locally (anymore)")
		Expect(sess.RemoveImage(ctx, localCanaryImage)).Error().NotTo(HaveOccurred())
		By("pulling the image")
		Expect(sess.PullImage(ctx, localCanaryImage,
			pull.WithRegistryAuth(magicauth),
			pull.WithOutput(timestamper.New(GinkgoWriter)))).To(Succeed())
		Expect(sess.HasImage(ctx, localCanaryImage)).To(BeTrue())
	})

})
