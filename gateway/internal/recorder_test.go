package internal_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/open-gateway/gateway/internal"
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
