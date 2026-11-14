// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// asyncWriter takes log lines without making the caller wait for the
// destination. One goroutine writes them through a buffer, which it flushes
// every flushEvery and on Close.
//
// When lines arrive faster than the destination takes them and the queue is
// full, new lines are dropped, so a slow disk never slows down requests. The
// number dropped is written as a warning line at the next flush.
type asyncWriter struct {
	lines      chan []byte
	quit       chan struct{}
	done       chan struct{}
	closeOnce  sync.Once
	dropped    atomic.Int64
	buf        *bufio.Writer
	flushEvery time.Duration
	now        func() time.Time
	writeErr   error
}

// writerOptions size an asyncWriter. Tests make them small.
type writerOptions struct {
	queue      int // lines that can wait before new ones are dropped
	bufferSize int // bytes collected before a write to the output
	flushEvery time.Duration
	now        func() time.Time
}

func newAsyncWriter(dest io.Writer, opts writerOptions) *asyncWriter {
	if opts.now == nil {
		opts.now = time.Now
	}
	w := &asyncWriter{
		lines:      make(chan []byte, opts.queue),
		quit:       make(chan struct{}),
		done:       make(chan struct{}),
		buf:        bufio.NewWriterSize(dest, opts.bufferSize),
		flushEvery: opts.flushEvery,
		now:        opts.now,
	}
	go w.run()
	return w
}

// Write queues a copy of p, or drops it if the queue is full. It never
// blocks and never fails, so logging cannot fail a request.
func (w *asyncWriter) Write(p []byte) (int, error) {
	line := make([]byte, len(p))
	copy(line, p)
	select {
	case w.lines <- line:
	default:
		w.dropped.Add(1)
	}
	return len(p), nil
}

// Close writes out the lines still queued, flushes, and stops the writer.
// Lines written after Close are dropped. It returns the first error the
// destination reported, if any.
func (w *asyncWriter) Close() error {
	w.closeOnce.Do(func() { close(w.quit) })
	<-w.done
	return w.writeErr
}

func (w *asyncWriter) run() {
	defer close(w.done)
	ticker := time.NewTicker(w.flushEvery)
	defer ticker.Stop()
	for {
		select {
		case line := <-w.lines:
			w.write(line)
		case <-ticker.C:
			w.flush()
		case <-w.quit:
			for {
				select {
				case line := <-w.lines:
					w.write(line)
				default:
					w.flush()
					return
				}
			}
		}
	}
}

func (w *asyncWriter) write(line []byte) {
	if _, err := w.buf.Write(line); err != nil && w.writeErr == nil {
		w.writeErr = err
	}
}

func (w *asyncWriter) flush() {
	if n := w.dropped.Swap(0); n > 0 {
		w.write([]byte(fmt.Sprintf(`{"time":%q,"level":"WARN","msg":"log lines dropped because the log output could not keep up","dropped":%d}`+"\n",
			w.now().UTC().Format(time.RFC3339Nano), n)))
	}
	if err := w.buf.Flush(); err != nil && w.writeErr == nil {
		w.writeErr = err
	}
}
