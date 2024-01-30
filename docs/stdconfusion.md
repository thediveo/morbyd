# Stdin, Stdout, Stderr, TTY, OpenStdin, ...

As we're working with containers and clients attaching to these containers'
input and output, things quickly get messy. Is this "stdout" the one of the
container? And where does it come from or goes to? Or is it the container's
"stdout" that the client want to receive when attaching?

Unfortunately, Docker's own API documentation is _very_ thin on this topic;
Docker's documentation people are probably in the same unedifying situation as
we are.

So here we are, trying to make sense of the several and confusing similar
configuration (`ContainerCreate`) and attachment (`ContainerAttach`) API
options.

## Docker Help

`docker help run` says...

| Option | Description |
| --- | --- |
| `-a, --attach list` | `Attach to STDIN, STDOUT or STDERR` |
| `-i, --interactive` | `Keep STDIN open even if not attached` |
| `-t, --tty` | `Allocate a pseudo-TTY` |
| `-d, --detach` | `Run container in background and print container ID` |

Please note that `--attach` and `--detach` are mutually exclusive.

## Docker `run` Variants

So how are `-a`, `-i`, and `-t` affecting certain `container.Config` fields?

| Options | `AttachStdin` | `AttachStdout`/ `AttachStderr` | `OpenStdin` | `StdinOnce` | `Tty` |
| --- | --- | --- | --- | --- | --- |
| | `false` | `true`/`true` | `false` | `false` | `false` |
| `--attach stdin` | `true` | `false`/`false` | `false` | `false` | `false` |
| `--attach stderr` | `false` | `false`/`true` | `false` | `false` | `false` |
| `-i` | **`true`** | `true`/`true` | **`true`** | **`true`** | `false` |
| `-it` | **`true`** | `true`/`true` | **`true`** | **`true`** | **`true`** |
| `-i --attach stderr` | **`true`** | `false`/**`true`** | **`true`** | **`true`** | `false` |
| `-d` | `false` | `false`/`false` | **`true`** | `false` | `false` |

- `-i` sets `AttachStdin` to `true`, as well as `OpenStdin` to `true` and also
  `StdinOnce` to `true`. Additionally, `-it` also sets `AttachStdout` and
  `AttachStderr` to `true`, but this can be overridden by using `--attach` to
  specify which of `stdout` and `stderr` should be attached.
- `-t` sets `Tty` to `true` – that's the simplest, straight one-to-one mapping.
- `-d` sets `OpenStdin` to `true`, but nothing else.

## Container Side (`ContainerCreate`)

`-i` sets up the **container's** `stdin`, `stdout`/`stderr`: however, containers
_always_ have their `stdout`/`stderr` open. In order to also provide an open
`stdin` (even if this currently has no client attached), `OpenStdin` must be
`true`.

Below the CLI surface, `AttachStdin`, `AttachStdout`, and `AttachStderr` in the
container's configuration control whether Docker assigns (when `false`)
`/dev/null` or otherwise either a fifo/pipe or (pseudo) TTY (when throwing in
`-t`).

Thus, `OpenStdin` controls whether a container has an open `stdin` at all; and
`StdinOnce` as `true` then closes the container's `stdin` after the first client
detaches.

## Client Side (`ContainerAttach`)

`--attach` wires up the **client's** `stdin`, `stdout`, and `stderr` with the
"remote" container's `stdin`, `stdout`, and `stderr`.

## `TTY`

The `Tty` field basically controls whether the container gets fifo/pipes or
`/dev/pts`.

## References

- [Understanding docker run --attach option](https://forums.docker.com/t/understanding-docker-run-attach-option/134337) (Docker Community Forum)
