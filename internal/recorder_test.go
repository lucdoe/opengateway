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

package internal_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal"
	"github.com/stretchr/testify/assert"
)

func TestResponseRecorder(t *testing.T) {
	mockResponse := httptest.NewRecorder()

	recorder := internal.NewResponseRecorder(mockResponse)

	statusCode := http.StatusNotFound
	recorder.WriteHeader(statusCode)
	assert.Equal(t, statusCode, recorder.StatusCode, "Status code should match")
	assert.Equal(t, statusCode, mockResponse.Code, "Underlying ResponseWriter status code should match")

	testData := []byte("Hello, world!")
	written, err := recorder.Write(testData)
	assert.Nil(t, err, "No error should occur on write")
	assert.Equal(t, len(testData), written, "Number of bytes written should match")
	assert.Equal(t, testData, recorder.Body.Bytes(), "Data in buffer should match written data")
	assert.Equal(t, testData, mockResponse.Body.Bytes(), "Data in underlying ResponseWriter should match")

	newResponse := httptest.NewRecorder()
	recorder.CopyBody(newResponse)
	assert.Equal(t, testData, newResponse.Body.Bytes(), "Copied data should match the original data")
}

func TestResponseRecorderDefaultStatusCode(t *testing.T) {
	mockResponse := httptest.NewRecorder()
	recorder := internal.NewResponseRecorder(mockResponse)

	testData := []byte("Test data")
	_, err := recorder.Write(testData)

	assert.Nil(t, err, "No error should occur on write")
	assert.Equal(t, http.StatusOK, recorder.StatusCode, "Default status code should be http.StatusOK when no status code is explicitly set")
}

type errorResponseWriter struct {
	httptest.ResponseRecorder
}

func (erw *errorResponseWriter) Write(b []byte) (int, error) {
	return 0, fmt.Errorf("simulated write error")
}

func TestResponseRecorderWriteError(t *testing.T) {
	mockResponse := &errorResponseWriter{*httptest.NewRecorder()}
	recorder := internal.NewResponseRecorder(mockResponse)

	testData := []byte("Test data")
	_, err := recorder.Write(testData)

	assert.NotNil(t, err, "Error should occur on write")
	assert.Equal(t, "simulated write error", err.Error(), "Error message should match simulated error")
}
