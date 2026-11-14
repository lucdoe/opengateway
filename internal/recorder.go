// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bytes"
	"net/http"
)

type ResponseRecorder struct {
	http.ResponseWriter
	Body       bytes.Buffer
	StatusCode int
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w}
}

func (rr *ResponseRecorder) WriteHeader(code int) {
	rr.StatusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *ResponseRecorder) Write(b []byte) (int, error) {
	if rr.StatusCode == 0 {
		rr.StatusCode = http.StatusOK
	}
	numBytes, err := rr.Body.Write(b)
	if err != nil {
		return numBytes, err
	}
	return rr.ResponseWriter.Write(b)
}
