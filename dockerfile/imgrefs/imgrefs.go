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

package imgrefs

import (
	"context"
	"fmt"
	"maps"
	"net"
	"os"
	"strings"

	bkclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/moby/client"
	"github.com/thediveo/nonstd/sets"
	"github.com/thediveo/nonstd/xiter"

	"github.com/thediveo/morbyd/v2"
)

const dockerimagePrefix = "docker-image://"

// ImageReferences determines from the (via WithDockerfile) specified Dockerfile
// all referenced Docker image references, such as “alpine:latest@sha256:...”.
// It uses the buildkit integrated in the Docker daemon to parse and evaluate
// the Dockerfile, resolving build arguments as needed.
func ImageReferences(ctx context.Context, sess *morbyd.Session, opts ...Opt) ([]string, error) {
	iros := Options{
		Dockerfile: "Dockerfile",
	}
	for _, opt := range opts {
		if err := opt(&iros); err != nil {
			return nil, err
		}
	}

	dockerfileText, err := os.ReadFile(iros.Dockerfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read dockerfile: %w", err)
	}

	buildArgs := maps.Collect(xiter.Map2(maps.All(iros.BuildArgs),
		func(k string, v *string) (string, string) {
			if v == nil {
				return k, "" // keep shtumm about build args without value
			}
			return k, *v
		}))

	// see also: https://github.com/docker/buildx/blob/f30ef86d21e91400b6e645964b5151aae46b1402/driver/docker/driver.go#L66
	buildkitClnt, err := bkclient.New(ctx, "",
		bkclient.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return sess.Client().(client.APIClient).DialHijack(ctx, "/grpc", "h2c", nil)
		}),
		bkclient.WithSessionDialer(func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
			return sess.Client().(client.APIClient).DialHijack(ctx, "/session", proto, meta)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("creating buildkit session failed, reason: %w", err)
	}
	defer func() { _ = buildkitClnt.Close() }()

	result, err := dockerfile2llb.Dockerfile2LLB(ctx,
		dockerfileText,
		dockerfile2llb.ConvertOpt{
			Config: dockerui.Config{
				BuildArgs: buildArgs,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("parsing the dockerfile failed, reason: %w", err)
	}

	def, err := result.State.Marshal(ctx)
	if err != nil {
		return nil, fmt.Errorf("processing the result of parsing the dockerfile failed, reason: %w", err)
	}

	imageRefs := sets.New[string]()
	for _, def := range def.Def {
		var op pb.Op
		if err := op.UnmarshalVT(def); err != nil {
			continue
		}
		src := op.GetSource()
		if src == nil {
			continue
		}
		ident := src.GetIdentifier()
		if imgref, ok := strings.CutPrefix(ident, dockerimagePrefix); ok {
			imageRefs.Add(imgref)
		}
	}

	return imageRefs.Elements(), nil
}
