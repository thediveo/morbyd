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
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/safe"
	"github.com/thediveo/morbyd/session"
)

// Run a container and gather its output.
//
// We start by creating a session using [NewSession]. Because unit tests may
// crash and leave test containers and networks behind, we enable
// “auto-cleaning” using the [session.WithAutoCleaning] option, passing it a
// unique label. This label can be either a unique key (“KEY=”) or a unique
// key-value pair (“KEY=VALUE”); either form is allowed, depending on how you
// like to structure and label your test containers and networks. Auto-cleaning
// runs automatically directly after session creation (to remove any left-overs
// from a previous test run) and then again when calling [Session.Close].
//
// Next, [Session.Run] creates the container and then runs (in our example) a
// command that we supplied as part of the run configuration.
//
// Because running the container and gathering its output are asynchronous
// operations, we [Container.Wait] for the container to have terminated before
// we pick up its output.
//
// Note: [safe.Buffer] is a [bytes.Buffer] that is safe for concurrent use.
func ExampleSession_Run() {
	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)
	defer cancel()

	sess, err := morbyd.NewSession(ctx,
		session.WithAutoCleaning("test.morbyd=example.session.run"))
	if err != nil {
		panic(err)
	}
	defer sess.Close(ctx)

	var out safe.Buffer
	container, err := sess.Run(ctx,
		"busybox",
		run.WithCommand("/bin/sh", "-c", "echo \"Hellorld!\""),
		run.WithAutoRemove(),
		run.WithCombinedOutput(&out))
	if err != nil {
		panic(err)
	}

	_ = container.Wait(ctx)
	fmt.Print(out.String())
	// Output: Hellorld!
}
