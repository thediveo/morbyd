/*
Package morbyd is a thin layer on top of the standard Docker Go client to easily
build and run ephemeral test Docker images and containers, and run commands
inside containers. It especially hides the gory details of how to stream the
output, and optionally input, of containers and commands. Just [io.Writer] and
[io.Reader].

morbyd makes heavy use of [option functions] in order to help test writers to
get a grip on Docker's (slightly) excessive knobs-for-everything API design. It
neatly groups the many With...() options in packages, such as [run] for “run
container” and [exec] for “container execute”. This design avoids stuttering
option names that would otherwise clash across different API operations for
common configuration elements, such as names, labels, and options.

# Features of morbyd

At a glance:

  - testable examples for common tasks to get you quickly up and running.
  - option function design with extensive [Go Doc comments] that IDEs show upon
    option completion. No more pseudo option function “callbacks” that are
    none the better than passing the original Docker config type verbatim.
  - uses the [official Docker Go client] in order to benefit from its security
    fixes, functional upgrades, and all the other nice things to to get directly
    from upstream.
  - “auto-cleaning” that runs when creating a new test session and again at its
    end, removing all containers and networks especially tagged using
    [session.WithAutoCleaning] for the test.
  - uses [context.Context] throughout the whole module, especially integrating
    well with testing frameworks (such as [Ginkgo]) that support automatic
    unit test context creation.
  - extensive unit tests with large coverage.

# Trivia

The module name “morby” is an amalgation of “[Moby (Dock)]” and “morbid” –
ephemeral – test containers.

[option functions]: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
[Go Doc comments]: https://tip.golang.org/doc/comment
[official Docker Go client]: https://pkg.go.dev/github.com/docker/docker/client
[Ginkgo]: https://pkg.go.dev/github.com/onsi/ginkgo
[Moby (Dock)]: https://www.docker.com/blog/call-me-moby-dock/
*/
package morbyd
