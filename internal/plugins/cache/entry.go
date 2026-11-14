// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"encoding/binary"
	"net/http"
	"strings"
	"time"
)

// entryMagic starts every stored response and names the format version. A
// value without it, such as a JSON entry from an older gateway, is a miss.
const entryMagic = "OGC1"

// entry is one stored response.
//
// Stored as: entryMagic, the time it was stored as 8 bytes of big-endian
// Unix nanoseconds, the number of header names, then per name its length and
// bytes, its number of values and each value's length and bytes, and finally
// the body's length and the body as it is. Lengths and counts are uvarints.
// The body is not encoded, so a hit costs no decoding of it.
type entry struct {
	header   http.Header
	body     string
	storedAt time.Time
}

func encodeEntry(header http.Header, body []byte, storedAt time.Time) string {
	size := len(entryMagic) + 8 + binary.MaxVarintLen64 + len(body)
	for name, values := range header {
		size += 2*binary.MaxVarintLen64 + len(name)
		for _, value := range values {
			size += binary.MaxVarintLen64 + len(value)
		}
	}

	buf := make([]byte, 0, size)
	buf = append(buf, entryMagic...)
	buf = binary.BigEndian.AppendUint64(buf, uint64(storedAt.UnixNano()))
	buf = binary.AppendUvarint(buf, uint64(len(header)))
	for name, values := range header {
		buf = appendString(buf, name)
		buf = binary.AppendUvarint(buf, uint64(len(values)))
		for _, value := range values {
			buf = appendString(buf, value)
		}
	}
	buf = binary.AppendUvarint(buf, uint64(len(body)))
	buf = append(buf, body...)
	return string(buf)
}

func appendString(buf []byte, s string) []byte {
	buf = binary.AppendUvarint(buf, uint64(len(s)))
	return append(buf, s...)
}

// decodeEntry reads a value written by encodeEntry. Header names, values
// and the body are slices of value, not copies. It reports false for a
// value in another format, one that is cut short, or one with bytes left
// over.
func decodeEntry(value string) (entry, bool) {
	if !strings.HasPrefix(value, entryMagic) || len(value) < len(entryMagic)+8 {
		return entry{}, false
	}
	r := reader{s: value, i: len(entryMagic) + 8}
	var nanos uint64
	for _, b := range []byte(value[len(entryMagic):r.i]) {
		nanos = nanos<<8 | uint64(b)
	}
	storedAt := time.Unix(0, int64(nanos))

	names, ok := r.count()
	if !ok {
		return entry{}, false
	}
	header := make(http.Header, names)
	for n := 0; n < names; n++ {
		name, ok := r.string()
		if !ok {
			return entry{}, false
		}
		count, ok := r.count()
		if !ok {
			return entry{}, false
		}
		values := make([]string, count)
		for v := range values {
			if values[v], ok = r.string(); !ok {
				return entry{}, false
			}
		}
		header[name] = values
	}
	body, ok := r.string()
	if !ok || r.i != len(value) {
		return entry{}, false
	}
	return entry{header: header, body: body, storedAt: storedAt}, true
}

// reader walks a stored value. Every read checks the bounds, so a damaged
// value reports false instead of panicking or allocating without limit.
type reader struct {
	s string
	i int
}

func (r *reader) uvarint() (uint64, bool) {
	var x uint64
	for shift := uint(0); shift < 64; shift += 7 {
		if r.i >= len(r.s) {
			return 0, false
		}
		b := r.s[r.i]
		r.i++
		x |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return x, true
		}
	}
	return 0, false
}

// count reads a number of items still to come. Every item takes at least
// one byte, so a count above the bytes left is damage.
func (r *reader) count() (int, bool) {
	n, ok := r.uvarint()
	if !ok || n > uint64(len(r.s)-r.i) {
		return 0, false
	}
	return int(n), true
}

func (r *reader) string() (string, bool) {
	n, ok := r.count()
	if !ok {
		return "", false
	}
	s := r.s[r.i : r.i+n]
	r.i += n
	return s, true
}
