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

package jsonmsgs

import (
	"context"
	"io"
	"strings"

	"github.com/moby/moby/api/types/jsonstream"
	"github.com/thediveo/nonstd/xslices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSON messages", func() {

	It("panics when creating a Streamer from a nil reader", func() {
		Expect(func() { _ = New(nil) }).To(Panic())
	})

	It("closes multiple times", func() {
		s := New(io.NopCloser(strings.NewReader("foobar!")))
		Expect(s.Close()).To(Succeed())
		Expect(s.Close()).To(Succeed())
	})

	It("reads", func() {
		const magic = "foobar!"
		s := New(io.NopCloser(strings.NewReader(magic)))
		defer func() { _ = s.Close() }()
		var t strings.Builder
		for {
			var b [1]byte
			n, _ := s.Read(b[:])
			if n == 0 {
				break
			}
			t.WriteByte(b[0])
		}
		Expect(t.String()).To(Equal(magic))
	})

	When("iterating over streamed messages (not hams)", func() {

		It("iterates over messages until EOF", func(ctx context.Context) {
			s := New(io.NopCloser(strings.NewReader(`{"status":"foo"}{"status":"bar"}`)))
			defer func() { _ = s.Close() }()
			msgs, errs := xslices.Collect2(s.JSONMessages(ctx))
			Expect(errs).To(HaveEach(BeNil()))
			Expect(xslices.Map(msgs, func(msg jsonstream.Message) string { return msg.Status })).To(
				Equal([]string{"foo", "bar"}))
		})

		It("iterates reports errors", func(ctx context.Context) {
			s := New(io.NopCloser(strings.NewReader(`{status:"foo"}`)))
			defer func() { _ = s.Close() }()
			msgs, errs := xslices.Collect2(s.JSONMessages(ctx))
			Expect(errs[0]).To(HaveOccurred())
			Expect(msgs[0]).To(Equal(jsonstream.Message{}))
		})

		It("stops iterating", func(ctx context.Context) {
			s := New(io.NopCloser(strings.NewReader(`{"status":"foo"}{"status":"bar"}`)))
			defer func() { _ = s.Close() }()
			count := 0
			for range s.JSONMessages(ctx) {
				count++
				break
			}
			Expect(count).To(Equal(1))
		})

		It("stops iterating when cancelling the context", func(ctx context.Context) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			// Note: this is a slightly involved test due to the json decoder trying
			// to gobble as much as possible: if we simply cancel() then it will
			// still happily read all available messages that fit into its internal
			// buffer. So we cancel() and then trip the decoder up with an invalid
			// JSON object. In turn, the iterator will now check first for a
			// cancelled context and report that instead of a decoding error.
			s := New(io.NopCloser(strings.NewReader(`{"status":"foo"}{status:"bar"}`)))
			defer func() { _ = s.Close() }()
			cancel()
			_, errs := xslices.Collect2(s.JSONMessages(ctx))
			Expect(errs).To(HaveExactElements(BeNil(), MatchError("context canceled")))
		})

	})

	Context("when waiting for the stream to finish", func() {

		It("returns after all is said and done", func(ctx context.Context) {
			s := New(io.NopCloser(strings.NewReader(`{"status":"foo"}{"status":"bar"}`)))
			Expect(s.Wait(ctx)).To(Succeed())
		})

		It("returns early on a streaming error", func(ctx context.Context) {
			s := New(io.NopCloser(strings.NewReader(`{"status":"foo"}{status":"bar"}`)))
			Expect(s.Wait(ctx)).To(HaveOccurred())
		})

		It("returns early on an error message", func(ctx context.Context) {
			s := New(io.NopCloser(strings.NewReader(`{"status":"foo"}{"errorDetail":{"message":"quel malheur!"}}`)))
			Expect(s.Wait(ctx)).To(HaveOccurred())
		})

	})

})
