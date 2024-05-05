package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal/middleware"
)

func dummyHandlerWithErrCheck(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		if err.Error() == "http: request body too large" {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func TestBodyLimitHandler(t *testing.T) {
	limit10bytes := int64(10)

	handler := middleware.BodyLimitHandler(limit10bytes)(http.HandlerFunc(dummyHandlerWithErrCheck))

	textCases := []struct {
		name         string
		body         string
		wantStatus   int
		wantResponse string
	}{
		{
			name:       "Request Body Exactly at the Limit",
			body:       "1234567890",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Request Body Below the Limit",
			body:       "12345",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Boundary Testing - Below the Limit",
			body:       "123456789",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Boundary Testing - Above the Limit",
			body:       "12345678901",
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range textCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString(tt.body))

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("%s: expected status code %d, got %d", tt.name, tt.wantStatus, w.Code)
			}
		})
	}
}
