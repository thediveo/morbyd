// Copyright 2026 Harald Albrecht.
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

package jsonmsgs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"sync"

	"github.com/moby/moby/api/types/jsonstream"
)

// MessageStreamer abstracts Moby's API client client.ImagePullResponse and
// client.ImagePushResponse. Unfortunately, moby's client implementation is
// internal, so we have to somehow mock it here in order to use it our mock
// tests.
type MessageStreamer interface {
	io.ReadCloser
	JSONMessages(ctx context.Context) iter.Seq2[jsonstream.Message, error]
	Wait(ctx context.Context) error
}

// Streamer implements MessageStreamer.
type Streamer struct {
	rc    io.ReadCloser
	close func() error
}

var _ MessageStreamer = (*Streamer)(nil)

// New returns a new JSON message streamer; it panics when given a nil
// io.ReadCloser.
func New(rc io.ReadCloser) *Streamer {
	if rc == nil {
		panic("nil io.ReadCloser")
	}
	return &Streamer{
		rc:    rc,
		close: sync.OnceValue(rc.Close),
	}
}

// Read from the message streamer.
func (s Streamer) Read(p []byte) (n int, err error) {
	return s.rc.Read(p)
}

// Close the message streamer, returning any error. Close is idempotent, closing
// the io.ReaderCloser only once and returning any error on this first close on
// any later calls.
func (s Streamer) Close() error {
	return s.close()
}

// JSONMessages iterates over all messages from a stream.
func (s Streamer) JSONMessages(ctx context.Context) iter.Seq2[jsonstream.Message, error] {
	return func(yield func(jsonstream.Message, error) bool) {
		// When the passed context gets cancelled we then close the underlying
		// reader, which in turn will make the JSON decoding process fail, so it
		// returns to us with an error, given control back to us.
		unregister := context.AfterFunc(ctx, func() { _ = s.Close() })

		defer func() {
			unregister()
			_ = s.Close()
		}()

		dec := json.NewDecoder(s)
		for {
			var msg jsonstream.Message
			err := dec.Decode(&msg)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return // keep shtumm, that's it.
				}
				if err := ctx.Err(); err != nil {
					yield(jsonstream.Message{}, err)
					return
				}
				yield(jsonstream.Message{}, err)
				return
			}
			if !yield(msg, err) {
				return
			}
		}
	}
}

// Wait until the stream completes, returning an error if the context was
// cancelled, decoding failed, transport failed, or a JSON error message was received.
func (s Streamer) Wait(ctx context.Context) error {
	for msg, err := range s.JSONMessages(ctx) {
		if err != nil {
			return err
		}
		if msg.Error != nil {
			return msg.Error
		}
	}
	return nil
}
