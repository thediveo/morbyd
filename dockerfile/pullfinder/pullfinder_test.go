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

package pullfinder

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	bkclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/moby/client"
	"github.com/thediveo/nonstd/sets"

	"github.com/thediveo/morbyd/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("finding the required pulls", func() {

	It("finds the image references to be pulled", func(ctx context.Context) {
		morbydsess := Successful(morbyd.NewSession(ctx))
		DeferCleanup(morbydsess.Close)

		// see also: https://github.com/docker/buildx/blob/f30ef86d21e91400b6e645964b5151aae46b1402/driver/docker/driver.go#L66
		buildkitClnt := Successful(bkclient.New(ctx, "",
			bkclient.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				return morbydsess.Client().(client.APIClient).DialHijack(ctx, "/grpc", "h2c", nil)
			}),
			bkclient.WithSessionDialer(func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
				return morbydsess.Client().(client.APIClient).DialHijack(ctx, "/session", proto, meta)
			}),
		))
		DeferCleanup(buildkitClnt.Close)

		dockerfile := Successful(os.ReadFile("../_test/irefs/Dockerfile"))
		res := Successful(
			dockerfile2llb.Dockerfile2LLB(ctx, dockerfile, dockerfile2llb.ConvertOpt{
				Config: dockerui.Config{
					BuildArgs: map[string]string{
						"ALPINE_TAG": "latest",
						"GOLANG_TAG": "1-alpine",
					},
				},
			}))
		Expect(res).NotTo(BeNil())
		def := Successful(res.State.Marshal(ctx))
		Expect(def).NotTo(BeNil())

		imageRefs := sets.New[string]()
		for _, def := range def.Def {
			var op pb.Op
			Expect(op.UnmarshalVT(def)).To(Succeed())
			src := op.GetSource()
			if src == nil {
				continue
			}
			ident := src.GetIdentifier()
			if imgref, ok := strings.CutPrefix(ident, "docker-image://"); ok {
				imageRefs.Add(imgref)
			}
		}

		fmt.Printf("imagerefs: %v\n", imageRefs.Elements())
	})

})
