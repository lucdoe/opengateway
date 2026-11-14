// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package statuswriter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal/statuswriter"
	"github.com/stretchr/testify/assert"
)

func TestWriterRecordsTheStatusAndBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	w := statuswriter.New(rec)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("hello"))
	w.Write([]byte(" world"))

	assert.Equal(t, http.StatusCreated, w.Status())
	assert.Equal(t, int64(11), w.Bytes())
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "hello world", rec.Body.String())
}

func TestWriterTakesAWriteWithoutHeaderAsOK(t *testing.T) {
	w := statuswriter.New(httptest.NewRecorder())

	w.Write([]byte("ok"))

	assert.Equal(t, http.StatusOK, w.Status())
}

func TestWriterReportsNothingBeforeTheHandlerWrites(t *testing.T) {
	w := statuswriter.New(httptest.NewRecorder())

	assert.Zero(t, w.Status())
	assert.Zero(t, w.Bytes())
}

func TestWriterSkipsInformationalResponses(t *testing.T) {
	w := statuswriter.New(httptest.NewRecorder())

	w.WriteHeader(http.StatusEarlyHints)
	w.WriteHeader(http.StatusBadGateway)

	assert.Equal(t, http.StatusBadGateway, w.Status())
}

func TestWriterKeepsTheFirstFinalStatus(t *testing.T) {
	w := statuswriter.New(httptest.NewRecorder())

	w.WriteHeader(http.StatusNotFound)
	w.WriteHeader(http.StatusOK) // ignored by net/http as a superfluous call

	assert.Equal(t, http.StatusNotFound, w.Status())
}

func TestWriterCanBeFlushed(t *testing.T) {
	rec := httptest.NewRecorder()
	w := statuswriter.New(rec)

	w.Write([]byte("part"))

	assert.NoError(t, http.NewResponseController(w).Flush())
	assert.True(t, rec.Flushed)
}
