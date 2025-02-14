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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/thediveo/morbyd/moby"
	"github.com/thediveo/morbyd/session"
	mock "go.uber.org/mock/gomock"
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
			DoAndReturn(func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
				return wrapped.ContainerAttach(ctx, container, options)
			})
	}
	if !slices.Contains(withouts, "ContainerCreate") {
		rec.ContainerCreate(Any, Any, Any, Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
				return wrapped.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
			})
	}
	if !slices.Contains(withouts, "ContainerInspect") {
		rec.ContainerInspect(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
				return wrapped.ContainerInspect(ctx, containerID)
			})
	}
	if !slices.Contains(withouts, "ContainerKill") {
		rec.ContainerKill(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID, signal string) error {
				return wrapped.ContainerKill(ctx, containerID, signal)
			})
	}
	if !slices.Contains(withouts, "ContainerList") {
		rec.ContainerList(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
				return wrapped.ContainerList(ctx, options)
			})
	}
	if !slices.Contains(withouts, "ContainerPause") {
		rec.ContainerPause(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string) error {
				return wrapped.ContainerPause(ctx, containerID)
			})
	}
	if !slices.Contains(withouts, "ContainerRemove") {
		rec.ContainerRemove(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options container.RemoveOptions) error {
				return wrapped.ContainerRemove(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerRestart") {
		rec.ContainerRestart(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options container.StopOptions) error {
				return wrapped.ContainerRestart(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerStart") {
		rec.ContainerStart(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options container.StartOptions) error {
				return wrapped.ContainerStart(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerStop") {
		rec.ContainerStop(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, options container.StopOptions) error {
				return wrapped.ContainerStop(ctx, containerID, options)
			})
	}
	if !slices.Contains(withouts, "ContainerUnpause") {
		rec.ContainerUnpause(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string) error {
				return wrapped.ContainerUnpause(ctx, containerID)
			})
	}
	if !slices.Contains(withouts, "ContainerWait") {
		rec.ContainerWait(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
				return wrapped.ContainerWait(ctx, containerID, condition)
			})
	}

	if !slices.Contains(withouts, "ContainerExecCreate") {
		rec.ContainerExecCreate(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
				return wrapped.ContainerExecCreate(ctx, container, config)
			})
	}
	if !slices.Contains(withouts, "ContainerExecStart") {
		rec.ContainerExecStart(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, execID string, config types.ExecStartCheck) error {
				return wrapped.ContainerExecStart(ctx, execID, config)
			})
	}
	if !slices.Contains(withouts, "ContainerExecAttach") {
		rec.ContainerExecAttach(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
				return wrapped.ContainerExecAttach(ctx, execID, config)
			})
	}
	if !slices.Contains(withouts, "ContainerExecInspect") {
		rec.ContainerExecInspect(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
				return wrapped.ContainerExecInspect(ctx, execID)
			})
	}

	if !slices.Contains(withouts, "ImageBuild") {
		rec.ImageBuild(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
				return wrapped.ImageBuild(ctx, buildContext, options)
			})
	}
	if !slices.Contains(withouts, "ImageList") {
		rec.ImageList(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
				return wrapped.ImageList(ctx, options)
			})
	}
	if !slices.Contains(withouts, "ImagePull") {
		rec.ImagePull(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
				return wrapped.ImagePull(ctx, refStr, options)
			})
	}
	if !slices.Contains(withouts, "ImagePush") {
		rec.ImagePush(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, image string, options image.PushOptions) (io.ReadCloser, error) {
				return wrapped.ImagePush(ctx, image, options)
			})
	}
	if !slices.Contains(withouts, "ImageRemove") {
		rec.ImageRemove(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
				return wrapped.ImageRemove(ctx, imageID, options)
			})
	}
	if !slices.Contains(withouts, "ImageTag") {
		rec.ImageTag(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, source, target string) error {
				return wrapped.ImageTag(ctx, source, target)
			})
	}

	if !slices.Contains(withouts, "NetworkCreate") {
		rec.NetworkCreate(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error) {
				return wrapped.NetworkCreate(ctx, name, options)
			})
	}
	if !slices.Contains(withouts, "NetworkInspect") {
		rec.NetworkInspect(Any, Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error) {
				return wrapped.NetworkInspect(ctx, networkID, options)
			})
	}
	if !slices.Contains(withouts, "NetworkList") {
		rec.NetworkList(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, options types.NetworkListOptions) ([]types.NetworkResource, error) {
				return wrapped.NetworkList(ctx, options)
			})
	}
	if !slices.Contains(withouts, "NetworkRemove") {
		rec.NetworkRemove(Any, Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, networkID string) error {
				return wrapped.NetworkRemove(ctx, networkID)
			})
	}

	if !slices.Contains(withouts, "ServerVersion") {
		rec.ServerVersion(Any).AnyTimes().
			DoAndReturn(func(ctx context.Context, networkID string) (types.Version, error) {
				return wrapped.ServerVersion(ctx)
			})
	}

	return cl
}
