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
	io "io"
	"net/http"
	"time"

	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
)

// Run a container that publishes its http service on loopback and then query
// this service for the “magic” phrase.
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
// Next, [Session.Run] creates the container and then runs (in our example) the
// http server command that we supplied as part of the run configuration. We
// additionally publish this http service on the host's (IPv4) loopback
// interface, on an ephemeral port.
//
// We then pick up the ephemeral port number and send an HTTP GET request to it,
// checking for the proper successful answer.
func ExampleContainer_PublishedPort() {
	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)
	defer cancel()

	sess, err := morbyd.NewSession(ctx,
		session.WithAutoCleaning("test.morbyd=example.container.port"))
	if err != nil {
		panic(err)
	}
	defer sess.Close(ctx)

	container, err := sess.Run(ctx,
		"busybox",
		run.WithCommand("/bin/sh", "-c", `echo "DOH!" > index.html && httpd -f -p 1234`),
		run.WithAutoRemove(),
		run.WithPublishedPort("127.0.0.1:1234"))
	if err != nil {
		panic(err)
	}
	svcAddrPort := container.
		PublishedPort("1234").
		Any().UnspecifiedAsLoopback().String()
	req, err := http.NewRequest(
		http.MethodGet, "http://"+svcAddrPort+"/", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}
	defer resp.Body.Close()
	phrase, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", phrase)
	// Output: DOH!
}
