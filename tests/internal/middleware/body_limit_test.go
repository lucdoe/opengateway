package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucdoe/opengateway/internal/middleware"
)

func TestBodyLimitHandler(t *testing.T) {
	handler := middleware.BodyLimitHandler(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.String()

		if len(body) > 10 {
			t.Errorf("Expected body length <= 10, got %d", len(body))
		}
	}))

	req := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString("12345678901"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected %d, got %d", http.StatusRequestEntityTooLarge, w.Code)
	}
}
