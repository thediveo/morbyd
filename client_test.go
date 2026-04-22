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

// Use mockgen in source mode because our Moby client interface definition is
// for testing only, and mockgen cannot see it in "reflect mode".

package morbyd

import (
	"context"
	"io"
	"slices"

	"github.com/moby/moby/client"
	mock "go.uber.org/mock/gomock"

	"github.com/thediveo/morbyd/v2/moby"
	"github.com/thediveo/morbyd/v2/session"
)

// WithMockController wraps the Docker client in our mock, using the specified
// controller.
func WithMockController(ctrl *mock.Controller, withoutForwarding ...string) session.Opt {
	return func(o *session.Options) error {
		o.Wrapper = func(client moby.Client) moby.Client {
			return newWrappedClient(ctrl, client, withoutForwarding)
		}
		return nil
	}
}

// Any is an instance of gomock's Any matcher; as it is stateless, we can pass
// it around multiple times.
var Any = mock.Any()

// newWrappedClient returns a new MockClient and configures its flight recorder
// to forward all intercepted method calls to the wrapped real Docker client,
// with the calls any number of times.
func newWrappedClient(ctrl *mock.Controller, wrapped moby.Client, withouts []string) moby.Client {
	cl := NewMockClient(ctrl)
	rec := cl.EXPECT()

	rec.Close().DoAndReturn(func() error { return wrapped.Close() })

	if !slices.Contains(withouts, "ContainerAttach") {
		rec.ContainerAttach(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, container string, options client.ContainerAttachOptions) (client.ContainerAttachResult, error) {
				return wrapped.ContainerAttach(ctx, container, options)
			})
	}
	if !slices.Contains(withouts, "ContainerCreate") {
		rec.ContainerCreate(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
				return wrapped.ContainerCreate(ctx, options)
			})
	}
	if !slices.Contains(withouts, "ContainerInspect") {
		rec.ContainerInspect(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
				return wrapped.ContainerInspect(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerKill") {
		rec.ContainerKill(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerKillOptions) (client.ContainerKillResult, error) {
				return wrapped.ContainerKill(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerList") {
		rec.ContainerList(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
				return wrapped.ContainerList(ctx, options)
			})
	}
	if !slices.Contains(withouts, "ContainerPause") {
		rec.ContainerPause(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerPauseOptions) (client.ContainerPauseResult, error) {
				return wrapped.ContainerPause(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerRemove") {
		rec.ContainerRemove(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerRemoveOptions) (client.ContainerRemoveResult, error) {
				return wrapped.ContainerRemove(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerRename") {
		rec.ContainerRename(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerRenameOptions) (client.ContainerRenameResult, error) {
				return wrapped.ContainerRename(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerRestart") {
		rec.ContainerRestart(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerRestartOptions) (client.ContainerRestartResult, error) {
				return wrapped.ContainerRestart(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerStart") {
		rec.ContainerStart(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerStartOptions) (client.ContainerStartResult, error) {
				return wrapped.ContainerStart(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerStop") {
		rec.ContainerStop(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerStopOptions) (client.ContainerStopResult, error) {
				return wrapped.ContainerStop(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerUnpause") {
		rec.ContainerUnpause(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerUnpauseOptions) (client.ContainerUnpauseResult, error) {
				return wrapped.ContainerUnpause(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerWait") {
		rec.ContainerWait(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options client.ContainerWaitOptions) client.ContainerWaitResult {
				return wrapped.ContainerWait(ctx, containerID, options)
			})
	}

	if !slices.Contains(withouts, "ExecAttach") {
		rec.ExecAttach(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, execID string, options client.ExecAttachOptions) (client.ExecAttachResult, error) {
				return wrapped.ExecAttach(ctx, execID, options)
			})
	}
	if !slices.Contains(withouts, "ExecCreate") {
		rec.ExecCreate(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, container string, options client.ExecCreateOptions) (client.ExecCreateResult, error) {
				return wrapped.ExecCreate(ctx, container, options)
			})
	}
	if !slices.Contains(withouts, "ExecStart") {
		rec.ExecStart(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, execID string, options client.ExecStartOptions) (client.ExecStartResult, error) {
				return wrapped.ExecStart(ctx, execID, options)
			})
	}
	if !slices.Contains(withouts, "ExecInspect") {
		rec.ExecInspect(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, execID string, options client.ExecInspectOptions) (client.ExecInspectResult, error) {
				return wrapped.ExecInspect(ctx, execID, options)
			})
	}

	if !slices.Contains(withouts, "ImageBuild") {
		rec.ImageBuild(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, buildContext io.Reader, options client.ImageBuildOptions) (client.ImageBuildResult, error) {
				return wrapped.ImageBuild(ctx, buildContext, options)
			})
	}
	if !slices.Contains(withouts, "ImageInspect") {
		rec.ImageInspect(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, imageID string, inspectOpts ...client.ImageInspectOption) (client.ImageInspectResult, error) {
				return wrapped.ImageInspect(ctx, imageID, inspectOpts...)
			})
	}
	if !slices.Contains(withouts, "ImageList") {
		rec.ImageList(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options client.ImageListOptions) (client.ImageListResult, error) {
				return wrapped.ImageList(ctx, options)
			})
	}
	if !slices.Contains(withouts, "ImagePull") {
		rec.ImagePull(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, refStr string, options client.ImagePullOptions) (client.ImagePullResponse, error) {
				return wrapped.ImagePull(ctx, refStr, options)
			})
	}
	if !slices.Contains(withouts, "ImagePush") {
		rec.ImagePush(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, image string, options client.ImagePushOptions) (client.ImagePushResponse, error) {
				return wrapped.ImagePush(ctx, image, options)
			})
	}
	if !slices.Contains(withouts, "ImageRemove") {
		rec.ImageRemove(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, imageID string, options client.ImageRemoveOptions) (client.ImageRemoveResult, error) {
				return wrapped.ImageRemove(ctx, imageID, options)
			})
	}
	if !slices.Contains(withouts, "ImageTag") {
		rec.ImageTag(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options client.ImageTagOptions) (client.ImageTagResult, error) {
				return wrapped.ImageTag(ctx, options)
			})
	}

	if !slices.Contains(withouts, "NetworkCreate") {
		rec.NetworkCreate(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, name string, options client.NetworkCreateOptions) (client.NetworkCreateResult, error) {
				return wrapped.NetworkCreate(ctx, name, options)
			})
	}
	if !slices.Contains(withouts, "NetworkInspect") {
		rec.NetworkInspect(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, networkID string, options client.NetworkInspectOptions) (client.NetworkInspectResult, error) {
				return wrapped.NetworkInspect(ctx, networkID, options)
			})
	}
	if !slices.Contains(withouts, "NetworkList") {
		rec.NetworkList(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options client.NetworkListOptions) (client.NetworkListResult, error) {
				return wrapped.NetworkList(ctx, options)
			})
	}
	if !slices.Contains(withouts, "NetworkRemove") {
		rec.NetworkRemove(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, networkID string, options client.NetworkRemoveOptions) (client.NetworkRemoveResult, error) {
				return wrapped.NetworkRemove(ctx, networkID, options)
			})
	}

	if !slices.Contains(withouts, "ServerVersion") {
		rec.ServerVersion(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options client.ServerVersionOptions) (client.ServerVersionResult, error) {
				return wrapped.ServerVersion(ctx, options)
			})
	}

	return cl
}
