// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecorderPassesTheResponseOnAndKeepsTheBody(t *testing.T) {
	w := httptest.NewRecorder()
	rec := &recorder{ResponseWriter: w}

	rec.WriteHeader(http.StatusNotFound)
	n, err := rec.Write([]byte("hello"))

	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, http.StatusNotFound, rec.status)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "hello", rec.body.String())
	assert.Equal(t, "hello", w.Body.String())
}

func TestRecorderDefaultsTo200(t *testing.T) {
	rec := &recorder{ResponseWriter: httptest.NewRecorder()}

	rec.Write([]byte("x"))

	assert.Equal(t, http.StatusOK, rec.status)
}

func TestRecorderStopsKeepingTheBodyOverTheLimit(t *testing.T) {
	w := httptest.NewRecorder()
	rec := &recorder{ResponseWriter: w}

	rec.Write(make([]byte, maxBodyBytes))
	assert.False(t, rec.tooBig, "exactly the limit is kept")
	rec.Write([]byte("x"))

	assert.True(t, rec.tooBig)
	assert.Zero(t, rec.body.Len(), "the partial copy is dropped")
	assert.Equal(t, maxBodyBytes+1, w.Body.Len(), "the client still gets everything")
}

type failingWriter struct{ httptest.ResponseRecorder }

func (*failingWriter) Write([]byte) (int, error) { return 0, errors.New("client went away") }

func TestRecorderReturnsWriteErrors(t *testing.T) {
	rec := &recorder{ResponseWriter: &failingWriter{*httptest.NewRecorder()}}

	_, err := rec.Write([]byte("x"))

	assert.EqualError(t, err, "client went away")
}

func TestRecorderUnwrapsToTheClientWriter(t *testing.T) {
	w := httptest.NewRecorder()

	assert.Same(t, w, (&recorder{ResponseWriter: w}).Unwrap())
}
