// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"bytes"
	"net/http"
)

// maxBodyBytes is the largest body the cache keeps. Larger responses are
// still delivered in full, just not stored.
const maxBodyBytes = 1 << 20

// recorder passes the response on to the client and keeps a copy of the body
// for the cache, up to maxBodyBytes.
type recorder struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
	// tooBig or streamed responses must not be stored.
	tooBig   bool
	streamed bool
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	if !r.tooBig && r.body.Len()+len(b) <= maxBodyBytes {
		r.body.Write(b)
	} else {
		r.tooBig = true
		r.body = bytes.Buffer{}
	}
	return r.ResponseWriter.Write(b)
}

// FlushError sends what is written so far, as streamed responses such as
// server-sent events need. A flushed response is not stored: it is a stream,
// not a document.
func (r *recorder) FlushError() error {
	r.streamed = true
	return http.NewResponseController(r.ResponseWriter).Flush()
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *recorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
