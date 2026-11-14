// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// syncBuffer is a bytes.Buffer that can be read while the writer goroutine
// writes to it.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// stalledWriter blocks every write until release is closed, like a disk that
// has stopped answering.
type stalledWriter struct {
	release chan struct{}
	syncBuffer
}

func (s *stalledWriter) Write(p []byte) (int, error) {
	<-s.release
	return s.syncBuffer.Write(p)
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) { return 0, errors.New("disk full") }

func TestAsyncWriterWritesLinesInOrderOnClose(t *testing.T) {
	var out syncBuffer
	w := newAsyncWriter(&out, writerOptions{queue: 100, bufferSize: 4096, flushEvery: time.Hour})

	for i := 0; i < 50; i++ {
		w.Write([]byte("line " + strconv.Itoa(i) + "\n"))
	}
	require.NoError(t, w.Close())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.Len(t, lines, 50)
	for i, line := range lines {
		assert.Equal(t, "line "+strconv.Itoa(i), line)
	}
}

func TestAsyncWriterFlushesOnTheInterval(t *testing.T) {
	var out syncBuffer
	w := newAsyncWriter(&out, writerOptions{queue: 100, bufferSize: 4096, flushEvery: 20 * time.Millisecond})
	t.Cleanup(func() { w.Close() })

	w.Write([]byte("hello\n"))

	assert.Eventually(t, func() bool { return out.String() == "hello\n" }, time.Second, 5*time.Millisecond)
}

func TestAsyncWriterCopiesWhatItIsGiven(t *testing.T) {
	var out syncBuffer
	w := newAsyncWriter(&out, writerOptions{queue: 10, bufferSize: 4096, flushEvery: time.Hour})

	line := []byte("first\n")
	w.Write(line)
	copy(line, "XXXXX\n") // slog reuses its buffer after Write returns
	require.NoError(t, w.Close())

	assert.Equal(t, "first\n", out.String())
}

// A disk that stops answering must not hold up requests.
func TestAsyncWriterDropsLinesInsteadOfBlocking(t *testing.T) {
	out := &stalledWriter{release: make(chan struct{})}
	// A buffer smaller than one line sends each line straight to the output,
	// so the first line already stalls the writer goroutine.
	w := newAsyncWriter(out, writerOptions{queue: 4, bufferSize: 1, flushEvery: time.Hour})

	started := time.Now()
	for i := 0; i < 100; i++ {
		n, err := w.Write([]byte("line\n"))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
	}
	assert.Less(t, time.Since(started), 100*time.Millisecond, "writes return at once while the output is stalled")

	close(out.release)
	require.NoError(t, w.Close())

	written := out.String()
	kept := strings.Count(written, "line\n")
	assert.Less(t, kept, 100)
	assert.Contains(t, written, `"level":"WARN","msg":"log lines dropped because the log output could not keep up","dropped":`+strconv.Itoa(100-kept)+`}`)
}

func TestAsyncWriterDropsLinesAfterClose(t *testing.T) {
	var out syncBuffer
	w := newAsyncWriter(&out, writerOptions{queue: 10, bufferSize: 4096, flushEvery: time.Hour})
	require.NoError(t, w.Close())

	_, err := w.Write([]byte("late\n"))

	assert.NoError(t, err)
	assert.NoError(t, w.Close(), "closing twice is fine")
	assert.Empty(t, out.String())
}

func TestAsyncWriterReportsTheOutputsError(t *testing.T) {
	w := newAsyncWriter(failingWriter{}, writerOptions{queue: 10, bufferSize: 1, flushEvery: time.Hour})

	w.Write([]byte("line\n")) // bigger than the buffer, so it goes straight out

	assert.EqualError(t, w.Close(), "disk full")
}

func TestAsyncWriterReportsAFailedFlush(t *testing.T) {
	w := newAsyncWriter(failingWriter{}, writerOptions{queue: 10, bufferSize: 4096, flushEvery: time.Hour})

	w.Write([]byte("line\n")) // fits the buffer; the flush fails

	assert.EqualError(t, w.Close(), "disk full")
}
