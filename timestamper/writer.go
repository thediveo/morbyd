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
	"io"
	"time"
	"unicode/utf8"
)

// DefaultFormat for the date and time prepended to each line of a TimeWriter.
const DefaultFormat = "01/02/06 15:04:05.000"

// Writer is an io.Writer preceding each line in the output with the current
// time and date. See also [time.Format].
type Writer struct {
	w       io.Writer
	stampit bool
	format  string
}

// New returns a new Writer object that time stamps each line of output and uses
// the default date and time format "01/02/06 15:04:05.999". See also
// [time.Format].
func New(w io.Writer) *Writer {
	return NewWithFormat(w, DefaultFormat)
}

// NewWithFormat returns a new Writer object hat time stamps each line of output
// and uses the specified date and time format.
func NewWithFormat(w io.Writer, format string) *Writer {
	return &Writer{
		w:       w,
		stampit: true,
		format:  format,
	}
}

// Write writes len(p) bytes from p to the underlying data stream, preceding
// each line (separated by '\n') with the current time and date.
func (t *Writer) Write(p []byte) (int, error) {
	nowtext := []byte(time.Now().Format(t.format) + " ")
	n := 0
	// We basically write the data given to us in chunks, inserting time and
	// date after each line terminator '\n' seen.
	for {
		if t.stampit {
			t.stampit = false
			_, err := t.w.Write(nowtext)
			if err != nil {
				return n, err
			}
		}
		// Find the first newline rune in the yet remaining p.
		idx := 0
		for idx < len(p) {
			r, size := utf8.DecodeRune(p[idx:])
			idx += size
			if r == '\n' {
				t.stampit = true
				break
			}
		}
		if idx >= len(p) {
			// idx points past the end of the remaining p, so we simply write
			// this remaining p and are good for now.
			nw, err := t.w.Write(p)
			return n + nw, err
		}
		// idx points to the next rune following the newline, so we write this
		// line and then rinse and repeat.
		//
		// https://pkg.go.dev/io#Writer: "Write must return a non-nil error if
		// it returns n < len(p)."
		nw, err := t.w.Write(p[:idx])
		n += nw
		if err != nil {
			return n, err
		}
		p = p[idx:]
	}
}
