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

package morbyd_test

import (
	"context"
	"fmt"
	"time"

	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/exec"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
)

// Execute a command inside a running container using [Container.Exec].
//
// In this example, we start by creating a session using [NewSession]. Because
// unit tests may crash and leave test containers and networks behind, we enable
// “auto-cleaning” using the [session.WithAutoCleaning] option, passing it a
// unique label. This label can be either a unique key (“KEY=”) or a unique
// key-value pair (“KEY=VALUE”); either form is allowed, depending on how you
// like to structure and label your test containers and networks. Auto-cleaning
// runs automatically directly after session creation (to remove any left-overs
// from a previous test run) and then again when calling [Session.Close].
//
// Next, we start a container using [Session.Run] where this container simply
// sits idle in a sleep loop (so that the idling shell process reacts more
// quickly to SIGTERMs).
//
// Then, we run a new command inside this container and pick up its output.
// Finally, we wind everything down.
//
// Note: [safe.Buffer] is a [bytes.Buffer] that is safe for concurrent use.
func ExampleContainer_Exec() {
	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)
	defer cancel()

	sess, err := morbyd.NewSession(ctx,
		session.WithAutoCleaning("test.morbyd=example.container.exec"))
	if err != nil {
		panic(err)
	}
	defer sess.Close(ctx)

	container, err := sess.Run(ctx,
		"busybox",
		run.WithCommand("/bin/sh", "-c", "trap 'exit 1' TERM; while true; do sleep 1; done"),
		run.WithAutoRemove())
	if err != nil {
		panic(err)
	}
	defer container.Kill(context.Background()) // just to be sure

	var out safe.Buffer
	exec, err := container.Exec(ctx,
		exec.Command("/bin/echo", "Hellorld! from exec"),
		exec.WithCombinedOutput(&out))
	if err != nil {
		panic(err)
	}
	exitcode, _ := exec.Wait(ctx)
	fmt.Printf("command exited with code %d\n", exitcode)
	container.Stop(ctx)
	fmt.Println(out.String())
	// Output: command exited with code 0
	// Hellorld! from exec
}
