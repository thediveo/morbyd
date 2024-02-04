<img alt="lxkns logo" align="right" width="200" src="docs/morbyd.png">

# `morbyd`

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/morbyd)
[![License](https://img.shields.io/github/license/thediveo/morbyd)](https://img.shields.io/github/license/thediveo/morbyd)
![build and test](https://github.com/thediveo/morbyd/workflows/build%20and%20test/badge.svg?branch=master)
![Coverage](https://img.shields.io/badge/Coverage-95.4%25-brightgreen)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/morbyd)](https://goreportcard.com/report/github.com/thediveo/morbyd)

`morbyd` is a thin layer on top of the standard Docker Go client to easily build
and run throw-away test Docker images and containers. And to run commands inside
these containers. In particular, `morbyd` hides the gory details of how to
stream the output, and optionally input, of container and commands via Dockers
API. You just use your `io.Writer`s and `io.Reader`s, for instance, to reason
about the expected output.

This module makes heavy use of [option
functions](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis).
So you can quickly get a grip on Docker's _slightly_ excessive
knobs-for-everything API design. `morbyd` neatly groups the many `With...()`
options in packages, such as `run` for "_run_ container" and `exec` for
"container *exec*ute". This design avoids stuttering option names that would
otherwise clash across different API operations for common configuration
elements, such as names, labels, and options.

## Features of morbyd

  - testable examples for common tasks to get you quickly up and running. Please
    see the [package
    documentation](https://pkg.go.dev/github.com/thediveo/morbyd).

  - option function design with extensive [Go Doc
    comments](https://tip.golang.org/doc/comment) that IDEs show upon option
    completion. No more pseudo option function "callbacks" that are none the
    better than passing the original Docker config type verbatim.

  - uses the [official Docker Go
    client](https://pkg.go.dev/github.com/docker/docker/client) in order to
    benefit from its security fixes, functional upgrades, and all the other nice
    things to get directly from upstream.
  
  - “auto-cleaning” that runs when creating a new test session and again at its
    end, removing all containers and networks especially tagged using
    `session.WithAutoCleaning` for the test.
  
  - uses `context.Context` throughout the whole module, especially integrating
    well with testing frameworks (such as
    [Ginkgo](https://pkg.go.dev/github.com/onsi/ginkgo)) that support automatic
    unit test context creation.

  - extensive unit tests with large coverage. We even mock the Docker client in
    order to cover the "unhappy paths", also known as "error handling". _Those_
    paths most TDD apostles forget to preach about for _some_ reason.

## Trivia

The module name `morbyd` is an amalgation of ["_Moby_
(Dock)"](https://www.docker.com/blog/call-me-moby-dock/) and _morbid_ –
ephemeral – test containers.

## Usage

```go
package main

import (
    "context"

    "github.com/thediveo/morbyd"
    "github.com/thediveo/morbyd/exec"
    "github.com/thediveo/morbyd/run"
    "github.com/thediveo/morbyd/session"
)

func main() {
    ctx := context.TODO()
    // note: error handling left out for brevity
    //
    // note: enable auto-cleaning of left-over containers and
    // networks, both when creating the session as well as when
    // closing the session. Use a unique label either in form of
    // "key=" or "key=value".
    sess, _ := morbyd.NewSession(ctx, session.WithAutoCleaning("test.mytest="))
    defer sess.Close(ctx)

    cntr, _ := sess.Run(ctx, "busybox",
        run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
        run.WithAutRemove(),
        run.WithCombinedOutput(os.Stdout))
    defer cntr.Stop(ctx)

    cmd, _ := cntr.Exec(ctx,
        exec.WithCommand("/bin/sh", "-c", "echo \"Hellorld!\""),
        exec.WithCombinedOutput(os.Stdout))
    exitcode, _ := cmd.Wait(ctx)
}
```

## Alternatives

Why `morbyd` when there are already other much bigger and long-time
battle-proven tools for using Docker images and containers in Go tests?

- for years, [@ory/dockertest](https://github.com/ory/dockertest) has served me
  well. Yet I eventually hit its limitations hard: for instance, dockertest
  cannot handle Docker's `100 CONTINUE` API protocol upgrades, because of its
  own proprietary Docker client implementation. However, this functionatly is
  essential in streaming container and command output and input – and thus only
  allowing diagnosing tests. Such issues are unresponded and unfixed. In
  addition, having basically to pass functions for configuration of Docker data
  structures is repeating the job of option functions at each and every
  dockertest call site.
- [Testcontainers for Go](https://golang.testcontainers.org/) as a much larger
  solution with a steep learning curve as well as some automatically installing
  infrastructure – while I admire this design, it is difficult to understand
  what _exactly_ is happening. Better keep it simple.

## Supported Go Versions

`morbyd` supports versions of Go that are noted by the [Go release
policy](https://golang.org/doc/devel/release.html#policy), that is, major
versions _N_ and _N_-1 (where _N_ is the current major version).

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Copyright and License

`morbyd` is Copyright 2024 Harald Albrecht, and licensed under the Apache
License, Version 2.0.

The package `github.com/thediveo/morbyd/run/dockercli` is [Copyright 2013-2017
Docker, Inc.](https://github.com/moby/moby/blob/v25.0.1/LICENSE) and licensed
under the Apache License Version 2.0, with the elements listed below coming from
the [github.com/docker/cli](https://github.com/docker/cli) module in order to
work around import dependency versioning problems due to `@docker/cli` using a
managed `vendor/` directory, but not providing a `go.mod` and the associated
guarantees:
- [opts/mount.go](https://github.com/docker/cli/blob/v25.0.1/opts/mount.go);
  removed the logrus dependency.
- [opts/network.go](https://github.com/docker/cli/blob/v25.0.1/opts/network.go)
- a subset of
  [cli/compose/types/types.go](https://github.com/docker/cli/blob/v25.0.1/cli/compose/types/types.go):
  type `ServiceVolumeConfig`, with direct dependency types `ServiceVolumeBind`,
  `ServiceVolumeVolume`, `ServiceVolumeTmpfs`, and `ServiceVolumeCluster`.
- [cli/compose/loader/volume.go](https://github.com/docker/cli/blob/v25.0.1/cli/compose/loader/volume.go)
