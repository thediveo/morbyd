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
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	bkcontrol "github.com/moby/buildkit/api/services/control"
	bkclient "github.com/moby/buildkit/client"
	bksession "github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/go-archive"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/jsonmessage"
	"github.com/moby/patternmatcher/ignorefile"
	"github.com/thediveo/nonstd/xatomic"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/thediveo/morbyd/v2/build"
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
//
// Note: using buildkit ([build.WithBuildKit]) currently is subject to
// limitations, most notably, registry authentication is not passed on to
// buildkit; please see also [API /build doesn't pass AuthConfig to BuildKit].
//
// [API /build doesn't pass AuthConfig to BuildKit]: https://github.com/moby/moby/issues/48112
func (s *Session) BuildImage(ctx context.Context, buildctxpath string, opts ...build.Opt) (id string, err error) {
	bios := build.Options{
		ImageBuildOptions: client.ImageBuildOptions{
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

	// We use an error wait group as in case of using buildkit we need to juggle
	// with multiple concurrent tasks that might error out sooner or later and
	// we need to then abort the other tasks; otherwise we have to wait for them
	// to all properly wind down before we can return our result.
	wg, ctx := errgroup.WithContext(ctx)
	sessionCtx, sessionDone := context.WithCancel(ctx)
	defer sessionDone()

	var statech chan *bkclient.SolveStatus // only for BuildKit
	closeStateCh := sync.OnceFunc(func() {
		if statech == nil {
			return
		}
		close(statech)
	})
	defer closeStateCh()

	if bios.Version == "2" {
		// the caller foolishly requests BuildKit and now hell breaks loose. In
		// order to build non-trivial Dockerfiles using the Docker
		// daemon-integrated buildkit successfully we need to create a new
		// buildkit session over a hijacked Docker API connection.
		buildkitSession, err := bksession.NewSession(sessionCtx, "")
		if err != nil {
			return "", fmt.Errorf("buildkit session creation failed, reason: %w", err)
		}
		defer func() { _ = buildkitSession.Close() }()

		wg.Go(func() error {
			// Aaaaarghhhhh!!! This is one of those pitch-dark long-running
			// function error reporting anti-pattern: in case the dialing fails,
			// buildkit.Session.Run returns early with an error. Otherwise it
			// tucks on and later always(!) returns nil. So when is the point of
			// no return where we won't get any error anymore? No-one knows.
			//
			// Where did I saw that anti-pattern before? Riiiight: the podman
			// native API!
			//
			// When the session enters the real "run" phase it won't ever return
			// anything other than nil, even if the context is cancelled or
			// times out, of the session gets closed.
			err := buildkitSession.Run(sessionCtx, func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
				return s.Client().(client.APIClient).DialHijack(ctx, "/session", proto, meta)
			})
			if sessionCtx.Err() != nil && ctx.Err() == nil {
				// if only our session was cancelled that is normal behavior, so
				// we should not report this as an error as we would otherwise
				// race with any other error return values from our waitgroup's
				// concurrent go routines.
				return nil
			}
			return err // ...but report "early" errors.
		})
		bios.SessionID = buildkitSession.ID()
		// of course, we want to leverage buildkit's client display and progress
		// UI, instead of being a dilettante. (Or is this instead spelled
		// "dilloitte" correctly?)
		//
		// Anyway, we create a display and run its processing loop in a
		// background go routine, feeding it status messages below as we receive
		// them via auxillary messages through the HTTP response body.
		//
		// This processing loop will cleanly terminate when the feeder closes
		// the status change channel. Cancelling the context will instead result
		// in an error ... which will be ignored by the waitgroup in case
		// another of its go routines has failed earlier.
		bkdisplay, err := progressui.NewDisplay(bios.Out, progressui.AutoMode)
		if err != nil {
			return "", fmt.Errorf("buildkit progress UI display creation failed, reason: %w", err)
		}
		statech = make(chan *bkclient.SolveStatus, 32)

		wg.Go(func() error {
			// UpdateFrom returns when either the state change channel has been
			// closed and drained, or when our context was done/cancelled ...
			// which happens when either one of the other go routines has failed
			// returning an error, or the context passed to our BuildImage
			// method has been done/cancelled.
			warnings, err := bkdisplay.UpdateFrom(ctx, statech)
			if err != nil {
				return err
			}
			// If this was a clean return from UpdateFrom, then render the
			// collected warnings, if any...
			for _, warning := range warnings {
				_, _ = bios.Out.Write([]byte(prettyPrintVertexWarning(warning)))
			}
			// ...and cancel our error waitgroup-derived sub context; now, as we
			// can only be done after the state change channel has been closed
			// there's only the above running buildkit session to terminate.
			// However, please note that we're still racing with the finishing
			// BuildImage/DisplayStream go routine returning any potential error
			// result. However, as we are on the path to success we're returning
			// a nil error and thus won't ever override the error return value
			// from Build/ImageDisplayStream. That leaves the buildkit session
			// go routine, so please see above.
			sessionDone()
			return nil
		})
	}

	var idval xatomic.Value[string] // never tickle the race detector
	wg.Go(func() error {
		// Now initiate the image build, feeding it our tar(r)ed build context
		// contents.
		resp, err := s.moby.ImageBuild(ctx, bios.Context, bios.ImageBuildOptions)
		if err != nil {
			return fmt.Errorf("image build failed, reason: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		err = jsonmessage.DisplayStream(resp.Body, bios.Out,
			jsonmessage.WithAuxCallback(func(auxmsg jsonstream.Message) {
				// buildkit messages are rather complex in that they are
				// protobuf-encoded and transmitted as aux messages with their
				// dedicated buildkit aux message ID. See also:
				// https://github.com/moby/moby/discussions/43788#discussioncomment-13291612
				// Digging deeper into the buildkit code base brings up the
				// progressui/client display stuff that feeds on status response
				// messages.
				if auxmsg.ID == "moby.buildkit.trace" {
					if auxmsg.Aux == nil {
						return
					}
					var bkpbmsg []byte
					if err := json.Unmarshal(*auxmsg.Aux, &bkpbmsg); err != nil {
						return
					}
					var status bkcontrol.StatusResponse
					if err := proto.Unmarshal(bkpbmsg, &status); err != nil {
						return
					}
					statech <- bkclient.NewSolveStatus(&status)
					return
				}
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
				idval.Store(aux.ID)
			}))
		closeStateCh()
		return err
	})

	err = wg.Wait()
	return idval.Load(), err
}

// readIgnorePatterns reads the file specified by “name” in .dockerignore
// format, returning the list of file patterns to ignore. In case of any error,
// it returns nil.
func readIgnorePatterns(name string) []string {
	f, err := os.Open(name)
	if err != nil {
		return nil
	}
	defer f.Close() //nolint:errcheck // any error is irrelevant at this point
	patterns, err := ignorefile.ReadAll(f)
	if err != nil {
		return nil
	}
	return patterns
}

func prettyPrintVertexWarning(warn bkclient.VertexWarning) string {
	const indentCount = 2
	var indent = strings.Repeat(" ", indentCount)

	var out bytes.Buffer

	out.WriteString("WARN: ")
	out.Write(warn.Short)
	out.WriteRune('\n')

	if warn.SourceInfo != nil && warn.SourceInfo.Filename != "" {
		out.WriteString(indent)
		out.WriteString("in: ")
		out.WriteString(warn.SourceInfo.Filename)
		out.WriteRune('\n')
	}

	for _, r := range warn.Range {
		out.WriteString(indent)
		out.WriteString("line ")
		fmt.Fprintf(&out, "%d:%d-%d:%d",
			r.Start.Line, r.Start.Character,
			r.End.Line, r.End.Character)
		out.WriteRune('\n')
	}

	if warn.URL != "" {
		out.WriteString(indent)
		out.WriteString("see also: ")
		out.WriteString(warn.URL)
		out.WriteRune('\n')
	}

	return out.String()
}
