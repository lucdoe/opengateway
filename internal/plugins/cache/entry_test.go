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

package cache

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var storedAt = time.Date(2026, 11, 14, 18, 0, 0, 123456789, time.UTC)

func sampleHeader() http.Header {
	return http.Header{
		"Content-Type":  {"application/json"},
		"Cache-Control": {"public", "max-age=60"},
		"Vary":          {"Accept-Encoding"},
		"X-Empty":       {""},
	}
}

func TestEntryRoundTrips(t *testing.T) {
	body := []byte(`{"results":[1,2,3]}` + "\x00\xff binary is fine")

	got, ok := decodeEntry(encodeEntry(sampleHeader(), body, storedAt))

	require.True(t, ok)
	assert.Equal(t, sampleHeader(), got.header)
	assert.Equal(t, string(body), got.body)
	assert.True(t, storedAt.Equal(got.storedAt))
}

func TestEntryRoundTripsAnEmptyResponse(t *testing.T) {
	got, ok := decodeEntry(encodeEntry(http.Header{}, nil, storedAt))

	require.True(t, ok)
	assert.Empty(t, got.header)
	assert.Empty(t, got.body)
}

func TestEntryKeepsTheBodyAsIs(t *testing.T) {
	body := bytes.Repeat([]byte("x"), 4096)

	value := encodeEntry(sampleHeader(), body, storedAt)

	assert.True(t, strings.HasSuffix(value, string(body)), "the body is stored raw, not base64")

	_, ok := decodeEntry(value + "extra")
	assert.False(t, ok, "bytes after the body mean the value is damaged")
	assert.Less(t, len(value), len(body)+200)
}

// Entries written by the gateway before this format are JSON; they must read
// as a miss, not as a broken response.
func TestDecodeEntryTreatsOtherFormatsAsMisses(t *testing.T) {
	oldJSON, err := json.Marshal(map[string]any{"header": sampleHeader(), "body": []byte("hi"), "stored_at": storedAt})
	require.NoError(t, err)

	for name, value := range map[string]string{
		"old JSON entry":  string(oldJSON),
		"empty":           "",
		"other version":   "OGC2" + strings.Repeat("\x00", 20),
		"magic only":      entryMagic,
		"no header count": entryMagic + "\x00\x00\x00\x00\x00\x00\x00\x01",
	} {
		t.Run(name, func(t *testing.T) {
			_, ok := decodeEntry(value)
			assert.False(t, ok)
		})
	}
}

func TestDecodeEntryRejectsEveryCutShortValue(t *testing.T) {
	value := encodeEntry(sampleHeader(), []byte("a body"), storedAt)

	for n := 0; n < len(value); n++ {
		_, ok := decodeEntry(value[:n])
		assert.False(t, ok, "cut at %d of %d bytes", n, len(value))
	}
}

func TestDecodeEntryRejectsCountsLargerThanTheValue(t *testing.T) {
	prefix := entryMagic + "\x00\x00\x00\x00\x00\x00\x00\x01"
	for name, rest := range map[string]string{
		"header count":       "\xff\xff\xff\xff\x0f",
		"name length":        "\x01\x7f",
		"value count":        "\x01\x01a\x7f",
		"value length":       "\x01\x01a\x01\x7f",
		"uvarint never ends": "\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff",
		"uvarint cut short":  "\x80",
	} {
		t.Run(name, func(t *testing.T) {
			_, ok := decodeEntry(prefix + rest)
			assert.False(t, ok)
		})
	}
}

func FuzzDecodeEntry(f *testing.F) {
	f.Add(encodeEntry(sampleHeader(), []byte("body"), storedAt))
	f.Add(entryMagic + "\x00\x00\x00\x00\x00\x00\x00\x01\x02")
	f.Add(`{"header":{},"body":"aGk="}`)
	f.Fuzz(func(t *testing.T, value string) {
		got, ok := decodeEntry(value) // must never panic
		if ok {
			again, ok := decodeEntry(encodeEntry(got.header, []byte(got.body), got.storedAt))
			require.True(t, ok)
			assert.Equal(t, got.body, again.body)
		}
	})
}
