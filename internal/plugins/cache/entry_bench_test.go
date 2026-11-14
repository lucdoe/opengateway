// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// jsonEntry is the format entries had before entryMagic, kept here as the
// baseline the benchmarks compare against.
type jsonEntry struct {
	Header   http.Header `json:"header"`
	Body     []byte      `json:"body"`
	StoredAt time.Time   `json:"stored_at"`
}

// A list response like the ones in the load tests: 4.1 KB of JSON and the
// headers a typical upstream sends.
var (
	benchBody   = []byte(`{"count":82,"results":[` + strings.Repeat(`{"name":"Person","height":"172","url":"https://example.test/api/people/1"},`, 55) + `{}]}`)
	benchHeader = http.Header{
		"Content-Type":   {"application/json"},
		"Content-Length": {"4105"},
		"Date":           {"Sat, 14 Nov 2026 18:00:00 GMT"},
		"Vary":           {"Accept-Encoding"},
		"Cache-Control":  {"public, max-age=60"},
	}
)

func BenchmarkEncode(b *testing.B) {
	b.Run("json", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			value, _ := json.Marshal(jsonEntry{Header: benchHeader, Body: benchBody, StoredAt: storedAt})
			_ = string(value)
		}
	})
	b.Run("binary", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = encodeEntry(benchHeader, benchBody, storedAt)
		}
	})
}

// A hit: decode the stored value and write the body out.
func BenchmarkHit(b *testing.B) {
	jsonValue, _ := json.Marshal(jsonEntry{Header: benchHeader, Body: benchBody, StoredAt: storedAt})
	jsonString := string(jsonValue)
	binaryValue := encodeEntry(benchHeader, benchBody, storedAt)
	b.Logf("stored size: json %d bytes, binary %d bytes, body %d bytes", len(jsonString), len(binaryValue), len(benchBody))

	b.Run("json", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var e jsonEntry
			_ = json.Unmarshal([]byte(jsonString), &e)
			io.Discard.Write(e.Body)
		}
	})
	b.Run("binary", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			e, _ := decodeEntry(binaryValue)
			io.WriteString(io.Discard, e.body)
		}
	})
}
