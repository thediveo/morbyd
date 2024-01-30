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

package timestamper

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func output(w io.Writer, s string) {
	GinkgoHelper()
	n, err := w.Write([]byte(s))
	Expect(err).NotTo(HaveOccurred())
	Expect(n).To(Equal(len(s)))
}

type errWriter struct {
	remaining int
}

func newErrWriter(n int) io.Writer {
	return &errWriter{remaining: n}
}

func (e *errWriter) Write(p []byte) (int, error) {
	e.remaining -= len(p)
	if e.remaining < 0 {
		return len(p) + e.remaining, errors.New("write limit exceeded")
	}
	return len(p), nil
}

var timestampRe = `\d{2}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}.\d{3} `

var _ = Describe("time(stamp) writer", func() {

	It("stamps two lines with a write boundaries", func() {
		var buff bytes.Buffer
		tw := New(&buff)

		output(tw, "first line\nsec")
		output(tw, "ond line\n")
		Expect(buff.String()).To(MatchRegexp(
			`(?m)^` + timestampRe + `first line\n` +
				timestampRe + `second line\n$`))

		output(tw, "third line")
		Expect(buff.String()).To(MatchRegexp(
			`(?m)^` + timestampRe + `first line\n` +
				timestampRe + `second line\n` +
				timestampRe + `third line$`))

		output(tw, "\nfo(u)rth line")
		Expect(buff.String()).To(MatchRegexp(
			`(?m)^` + timestampRe + `first line\n` +
				timestampRe + `second line\n` +
				timestampRe + `third line\n` +
				timestampRe + `fo\(u\)rth line$`))
	})

	It("handles downstream writer errors", func() {
		tw := NewWithFormat(newErrWriter((1+2)+(1+2)-1), "")
		n, err := tw.Write([]byte("A\nB\n"))
		Expect(n).To(Equal(2 + 1))
		Expect(err).To(HaveOccurred())

		tw = NewWithFormat(newErrWriter((1+2)+(1+1)), "")
		Expect(tw.Write([]byte("A"))).To(Equal(1))
		Expect(tw.Write([]byte("\nB"))).To(Equal(2))
		n, err = tw.Write([]byte("\nC"))
		Expect(n).To(Equal(0))
		Expect(err).To(HaveOccurred())

		tw = New(newErrWriter(2))
		n, err = tw.Write([]byte("A\nB\n"))
		Expect(n).To(Equal(0))
		Expect(err).To(HaveOccurred())
	})

})
