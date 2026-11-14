// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package logger is the logger plugin: one JSON line per request, written
// after the request has been answered.
package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

const (
	// queueLines is how many lines can wait for the output before new ones
	// are dropped.
	queueLines = 8192
	// bufferBytes is how much is collected before a write to the output.
	bufferBytes = 64 * 1024
	// flushInterval is the longest a written line waits in the buffer.
	flushInterval = time.Second
)

// Config says where the log goes.
type Config struct {
	// FilePath is the file lines are appended to. Empty means stdout.
	FilePath string
}

// Logger writes request lines as JSON. Lines are queued and written by a
// background goroutine; call Close to write out what is still queued.
type Logger struct {
	log  *slog.Logger
	out  *asyncWriter
	file io.Closer
}

// New opens the log output cfg names.
func New(cfg Config) (*Logger, error) {
	if cfg.FilePath == "" {
		return newLogger(os.Stdout, nil), nil
	}
	file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening log file %q: %w", cfg.FilePath, err)
	}
	return newLogger(file, file), nil
}

func newLogger(dest io.Writer, file io.Closer) *Logger {
	out := newAsyncWriter(dest, writerOptions{queue: queueLines, bufferSize: bufferBytes, flushEvery: flushInterval})
	return &Logger{
		log:  slog.New(slog.NewJSONHandler(out, nil)),
		out:  out,
		file: file,
	}
}

// Request logs one answered request.
func (l *Logger) Request(ctx context.Context, attrs ...slog.Attr) {
	l.log.LogAttrs(ctx, slog.LevelInfo, "request", attrs...)
}

// Close writes out the queued lines and closes the log file, if any.
func (l *Logger) Close() error {
	err := l.out.Close()
	if l.file != nil {
		err = errors.Join(err, l.file.Close())
	}
	return err
}
