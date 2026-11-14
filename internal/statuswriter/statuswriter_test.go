// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
