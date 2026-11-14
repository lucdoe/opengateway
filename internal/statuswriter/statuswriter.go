// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package statuswriter records what a handler answered, its status code and
// how many body bytes it wrote, without buffering the body.
package statuswriter

import "net/http"

// Writer passes everything on to the ResponseWriter it wraps and remembers
// the status and the number of body bytes.
type Writer struct {
	http.ResponseWriter
	status int
	bytes  int64
}

// New wraps w.
func New(w http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: w}
}

func (w *Writer) WriteHeader(code int) {
	// 1xx responses such as 103 Early Hints come before the real status.
	if w.status == 0 && code >= 200 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *Writer) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += int64(n)
	return n, err
}

// Status is the status code sent, or 0 if the handler has written nothing
// yet. A handler that returns without writing answers 200.
func (w *Writer) Status() int {
	return w.status
}

// Bytes is the number of body bytes written so far.
func (w *Writer) Bytes() int64 {
	return w.bytes
}

// Unwrap lets http.ResponseController reach the underlying writer, so
// streamed responses can still be flushed.
func (w *Writer) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
