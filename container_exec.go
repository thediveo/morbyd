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
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/thediveo/morbyd/exec"
)

// ExecSession represents a command running inside a container.
//
// Nota bene: the Docker API doesn't have an API endpoint for deleting
// executions when they're finished and we've picked up the results.
type ExecSession struct {
	ID        string     // command execution ID.
	Container *Container // container this command runs inside.

	// closes after the output stream from (and optionally our input stream to)
	// the command has been closed after the command has finally terminated.
	done chan struct{}
}

// Exec a command inside a container, using the specified command using
// [exec.Command](cmd, args...) and optional configuration information. It
// returns an *ExecSession object if successful, otherwise an error.
//
// Important: executing the command is a fully asynchronous process to the
// extend that the session returned might still in its startup phase with the
// command not yet being executed. [ExecSession.PID] can be re-appropriated to
// wait for the command have been started.
//
// Note: when using [exec.WithInput] make sure to close the input reader in
// order to not leak go routines handling the executed input and output streams
// in the background.
//
// Note: morbyd does not support executing detached commands, so we will always
// be attached to the executing command's input/output streams.
func (c *Container) Exec(ctx context.Context, cmd exec.Cmd, opts ...exec.Opt) (es *ExecSession, err error) {
	exopts := exec.Options{
		Conf: container.ExecOptions{
			Cmd: cmd,
		},
	}
	for _, opt := range opts {
		if err := opt(&exopts); err != nil {
			return nil, fmt.Errorf("cannot execute command in container %q/%s, reason: %w",
				c.Name, c.AbbreviatedID(), err)
		}
	}

	if exopts.Out == nil {
		exopts.Out = io.Discard
	}
	exopts.Conf.AttachStdout = true
	exopts.Conf.AttachStderr = true
	exopts.Conf.AttachStdin = exopts.In != nil

	// To quote from the Docker CLI
	// (https://github.com/docker/cli/blob/9e2615bc467fb4ec9a177049a6f2b4fbe5a20e65/cli/command/container/exec.go#L107):
	//
	//   "We need to check the tty _before_ we do the ContainerExecCreate, because
	//    otherwise if we error out we will leak execIDs on the server (and there's
	//    no easy way to clean those up). But also in order to make "not exist"
	//    errors take precedence we do a dummy inspect first."
	if _, err := c.Session.moby.ContainerInspect(ctx, c.ID); err != nil {
		return nil, fmt.Errorf("cannot execute into container %q/%s, reason: %w",
			c.Name, c.AbbreviatedID(), err)
	}

	execResp, err := c.Session.moby.ContainerExecCreate(ctx, c.ID, exopts.Conf)
	if err != nil {
		return nil, fmt.Errorf("cannot execute into container %q/%s, reason: %w",
			c.Name, c.AbbreviatedID(), err)
	}

	// now start executing the command and at the same time attach to its input
	// and output. Nota bene: the Docker Go client is confusing here, as there's
	// also a ContainerExecStart which also starts the exec but doesn't attach.
	attachResp, err := c.Session.moby.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty:         exopts.Conf.Tty,
		ConsoleSize: exopts.Conf.ConsoleSize,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot attach to exec session in container %q/%s, reason: %w",
			c.ID, c.AbbreviatedID(), err)
	}

	stdinDone := make(chan struct{})
	allDone := make(chan struct{})
	// Deal with the output stream...
	go func() {
		defer func() {
			<-stdinDone
			attachResp.Close()
			close(allDone)
		}()

		if exopts.Conf.Tty {
			// When using a TTY, only copy the single combined output stream
			// into the output stream specified in the run options.
			_, _ = io.Copy(exopts.Out, attachResp.Reader)
			return
		}
		// When NOT using a TTY, use Docker's own helper to demux the two
		// multiplexed streams into stdout and stderr writers.
		stderr := exopts.Err
		if stderr == nil {
			stderr = io.Discard
		}
		_, _ = stdcopy.StdCopy(exopts.Out, stderr, attachResp.Reader)
	}()
	// Deal with the input stream, where necessary.
	if exopts.In != nil {
		go func() {
			defer close(stdinDone)
			_, _ = io.Copy(attachResp.Conn, exopts.In)
		}()
	} else {
		close(stdinDone)
	}

	// At this point the command might not have actually been started in the
	// container, as we've found out the hard time in the early days of our unit
	// tests... but we return anyway, as the session is established but might
	// still fail asynchronously.
	exec := &ExecSession{
		ID:        execResp.ID,
		Container: c,
		done:      allDone,
	}
	return exec, nil
}

// PID returns the executing command's PID, or an error if the command has
// already terminated.
//
// Note: [Container.Run] can already return while the underlying Docker session
// for executing the command inside the container is still starting up. In this
// case, PID will wait until the executing command's PID becomes available or
// the passed context gets cancelled.
func (e *ExecSession) PID(ctx context.Context) (int, error) {
	for {
		inspRes, err := e.Container.Session.moby.ContainerExecInspect(ctx, e.ID)
		if err != nil {
			return 0, fmt.Errorf("cannot determine the PID of the excuting command, reason: %w",
				err)
		}
		if inspRes.Running && inspRes.Pid != 0 {
			return inspRes.Pid, nil
		}
		// We might end up being too early and thus getting our signals
		// crossed. If the output has already finished, then we're already
		// too late.
		select {
		case <-e.done:
			return 0, errors.New("command has already terminated")
		default:
		}
		// We're early, so let's take a tiny nap that can be cancelled at
		// any time via the context, in order to not go into Ludicrous
		// Speed.
		if err := Sleep(ctx, DefaultSleep); err != nil {
			return 0, err
		}
	}
}

// Done returns a channel that gets closed when the command has finished
// executing inside its container.
func (e *ExecSession) Done() chan struct{} {
	return e.done
}

// Wait for the command executed inside its container to finish, and then return
// the command's exit code. If the passed context gets cancelled or there is a
// problem picking up the command's exit code, Wait returns an error instead.
func (e *ExecSession) Wait(ctx context.Context) (exitcode int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-e.done:
		// fall through into happy path...
	}
	inspRes, err := e.Container.Session.moby.ContainerExecInspect(ctx, e.ID)
	if err != nil {
		return 0, fmt.Errorf("error fetching result code of executed command, reason: %w", err)
	}
	if inspRes.Running {
		return 0, fmt.Errorf("command execution still alive when it should not")
	}
	return inspRes.ExitCode, nil
}
