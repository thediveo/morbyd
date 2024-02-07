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
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/thediveo/morbyd/run"
	"golang.org/x/exp/maps"
)

// Run (create and start) a new container, using the referenced image and
// optional configuration information, returning a *Container object if
// successful. Otherwise, it returns an error without leaving behind any
// container.
//
// Additionally, Run attaches to the container's input and output streams which
// can be accessed using [run.WithInput], and either [run.WithCombinedOutput] or
// [run.WithDemuxedOutput].
//
// If the session has configured with labels, the new container inherits them.
// Use [run.ClearLabels] before [run.WithLabel] or [run.WithLabels] in order to
// remove any inherited labels first.
func (s *Session) Run(ctx context.Context, imageref string, opts ...run.Opt) (cntr *Container, err error) {
	// Our interpretation of a "container run" command differs in some aspect
	// from Docker's CLI "run" command implementation, see also here:
	// https://github.com/docker/cli/blob/f18a476b6d240bafbaefb65e51f837eab9e02b94/cli/command/container/run.go#L121

	copts := run.Options{
		Conf: container.Config{
			Image:  imageref,
			Labels: map[string]string{},
		},
	}
	maps.Copy(copts.Conf.Labels, s.opts.Labels) // inherit labels from session.
	for _, opt := range opts {
		if err := opt(&copts); err != nil {
			return nil, err
		}
	}

	if copts.Out == nil {
		copts.Out = io.Discard
	}

	// Pull the referenced image if it isn't already available locally.
	hasimg, err := s.HasImage(ctx, imageref)
	if err != nil {
		return nil, fmt.Errorf("cannot run image %s, reason: %w", imageref, err)
	}
	if !hasimg {
		if err := s.PullImage(ctx, imageref); err != nil {
			return nil, fmt.Errorf("cannot pull image, reason: %w", err)
		}
	}

	// Create the container; this doesn't start it yet.
	createResp, err := s.moby.ContainerCreate(ctx,
		&copts.Conf,
		&copts.Host,
		&copts.Net,
		nil,
		copts.Name)
	if err != nil {
		return nil, fmt.Errorf("cannot create container, reason: %w", err)
	}
	cntrID := createResp.ID

	// Whatever happens next, whenever on our way out of this method we see that
	// there's an error returned, then take the newly created container down
	// before we leave.
	defer func() {
		if err == nil {
			return
		}
		_ = s.moby.ContainerRemove(ctx, cntrID, container.RemoveOptions{
			Force: true,
		})
	}()

	// Now that the container is created but not yet started, let's attach to
	// the container's input and output streams.
	//
	// Now, when NOT using a TTY, the container's stdout and stderr are streamed
	// separately. But when using a TTY, we get the container's stdout+stderr
	// mixed together.
	attachResp, err := s.moby.ContainerAttach(ctx,
		cntrID, container.AttachOptions{
			Stream: true,
			Stdout: true,
			Stderr: true,
			Stdin:  copts.In != nil,
		})
	if err != nil {
		return nil, fmt.Errorf("cannot attach to container, reason: %w", err)
	}

	done := make(chan struct{})
	// Deal with the output stream...
	go func() {
		defer func() {
			<-done
			attachResp.Close()
		}()

		if copts.Conf.Tty {
			// When using a TTY, only copy the single combined output stream
			// into the output stream specified in the run options.
			_, _ = io.Copy(copts.Out, attachResp.Reader)
			return
		}
		// When NOT using a TTY, use Docker's own helper to demux the two
		// multiplexed streams into stdout and stderr writers.
		stderr := copts.Err
		if stderr == nil {
			stderr = io.Discard
		}
		_, _ = stdcopy.StdCopy(copts.Out, stderr, attachResp.Reader)
	}()
	// Deal with the input stream, where necessary.
	if copts.In != nil {
		go func() {
			defer close(done)
			_, _ = io.Copy(attachResp.Conn, copts.In)
		}()
	} else {
		close(done)
	}

	// Finally, we can try to start the container.
	if err := s.moby.ContainerStart(ctx, cntrID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("cannot start container, reason: %w", err)
	}
	details, err := s.moby.ContainerInspect(ctx, cntrID)
	if err != nil {
		return nil, fmt.Errorf("cannot inspect newly started container, reason: %w", err)
	}
	cntr = &Container{
		Name:    details.Name[1:],
		ID:      createResp.ID,
		Session: s,
		Details: details,
	}
	return cntr, nil
}
