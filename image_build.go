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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/buildkit/frontend/dockerfile/dockerignore"
	"github.com/thediveo/morbyd/build"
)

// BuildImage builds a container image using the specified build context and
// further build options. These build options are applied in the order they are
// provided, which allows modifying (or even nuking) the defaults when building
// an image.
//
// BuildImage returns the ID of the built image, or an error in case of build
// errors.
//
// Unless overridden using a build option, the following defaults apply:
//   - Dockerfile: "Dockerfile"
//   - Remove: true
//   - ForceRemove: true
//
// If no build process output writer has been specified using [build.WithOutput]
// any output (such as build steps, et cetera) will simply be discarded.
func (s *Session) BuildImage(ctx context.Context, buildctxpath string, opts ...build.Opt) (id string, err error) {
	bios := build.Options{
		ImageBuildOptions: types.ImageBuildOptions{
			Dockerfile:  "Dockerfile",
			Remove:      true,
			ForceRemove: true,
		},
	}
	for _, opt := range opts {
		if err := opt(&bios); err != nil {
			return "", err
		}
	}
	// In case no output writer was set, default to the discarding writer.
	if bios.Out == nil {
		bios.Out = io.Discard
	}
	// Tar up the files forming the build context, obeying the rules set down in
	// a .dockerignore where present.
	if bios.Context == nil {
		buildCtxTar, err := archive.TarWithOptions(buildctxpath,
			&archive.TarOptions{
				ExcludePatterns: readIgnorePatterns(filepath.Join(buildctxpath, ".dockerignore")),
			})
		if err != nil {
			return "", fmt.Errorf("cannot create build context, reason: %w", err)
		}
		bios.Context = buildCtxTar
	}
	// Now initiate the image build, feeding it our tar(r)ed build context
	// contents.
	resp, err := s.moby.ImageBuild(ctx, bios.Context, bios.ImageBuildOptions)
	if err != nil {
		return "", fmt.Errorf("image build failed, reason: %w", err)
	}
	defer resp.Body.Close()
	// https://stackoverflow.com/a/48579861 pointing to:
	// https://pkg.go.dev/github.com/moby/moby/pkg/jsonmessage?utm_source=godoc#DisplayJSONMessagesStream
	err = jsonmessage.DisplayJSONMessagesStream(resp.Body,
		bios.Out, 0, false,
		func(auxmsg jsonmessage.JSONMessage) {
			// Please note that the image ID is reported using an aux message
			// with its own embedded JSON message and not directly via an "ID"
			// JSON message.
			aux := struct {
				ID string `json:"ID"`
			}{}
			if err := json.Unmarshal(*auxmsg.Aux, &aux); err != nil || aux.ID == "" {
				return
			}
			// Pick up the image ID when it floats by ... and is non-zero.
			id = aux.ID
		})
	return id, err
}

// readIgnorePatterns reads the file specified by “name” in .dockerignore
// format, returning the list of file patterns to ignore. In case of any error,
// it returns nil.
func readIgnorePatterns(name string) []string {
	f, err := os.Open(name)
	if err != nil {
		return nil
	}
	defer f.Close()
	patterns, err := dockerignore.ReadAll(f)
	if err != nil {
		return nil
	}
	return patterns
}
