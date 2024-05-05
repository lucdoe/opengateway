package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lucdoe/opengateway/internal/middleware"
)

func TestGZIPHandler(t *testing.T) {
	sampleHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a test response"))
	})

	gzipHandler := middleware.GZIPHandler(sampleHandler)

	t.Run("GZIP Compression", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Header.Add("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.Header.Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding to be gzip, got %s", resp.Header.Get("Content-Encoding"))
		}

		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatalf("Expected gzip encoded response body, got error: %v", err)
		}
		defer reader.Close()

		body, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if !strings.Contains(string(body), "This is a test response") {
			t.Errorf("Response does not contain expected text, got: %s", string(body))
		}
	})

	t.Run("No GZIP Compression", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)

		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if ce := resp.Header.Get("Content-Encoding"); ce != "" {
			t.Errorf("Expected no Content-Encoding, got %s", ce)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if !bytes.Contains(body, []byte("This is a test response")) {
			t.Errorf("Response does not contain expected text, got: %s", string(body))
		}
	})
}
