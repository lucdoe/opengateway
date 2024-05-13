package logger_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/logger"
)

type MockFileWriter struct {
	Writer   *bytes.Buffer
	WriteErr error
}

func (m *MockFileWriter) Write(p []byte) (n int, err error) {
	if m.WriteErr != nil {
		return 0, m.WriteErr
	}
	return m.Writer.Write(p)
}

type MockFileOpener struct {
	File    logger.FileWriter
	OpenErr error
}

func (m *MockFileOpener) OpenFile(name string, flag int, perm os.FileMode) (logger.FileWriter, error) {
	if m.OpenErr != nil {
		return nil, m.OpenErr
	}
	return m.File, nil
}

type MockTimeProvider struct {
	FixedTime time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.FixedTime
}

func TestOSLoggerInfo(t *testing.T) {
	buffer := new(bytes.Buffer)
	mockWriter := &MockFileWriter{Writer: buffer}
	mockOpener := &MockFileOpener{File: mockWriter}
	mockTime := &MockTimeProvider{FixedTime: time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC)}
	errorBuffer := new(bytes.Buffer)

	cfg := logger.LoggerConfig{
		FilePath:     "testpath",
		FileWriter:   mockWriter,
		ErrOutput:    errorBuffer,
		TimeProvider: mockTime,
		FileOpener:   mockOpener,
	}

	l, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	msg, details := "test message", "test details"
	l.Info(msg, details)
	expectedLog := "2020-01-01 12:00:00 [INFO]: test message test details\n"
	if got := buffer.String(); got != expectedLog {
		t.Errorf("Expected log '%s', got '%s'", expectedLog, got)
	}

	mockWriter.WriteErr = errors.New("write failure")
	l.Info(msg, details)
	expectedErr := "Error writing to log file: write failure\n"
	if got := errorBuffer.String(); !strings.Contains(got, expectedErr) {
		t.Errorf("Expected error log to contain '%s', got '%s'", expectedErr, got)
	}

	cfgNoFileWriter := logger.LoggerConfig{
		FilePath:     "testpath",
		FileWriter:   nil,
		ErrOutput:    errorBuffer,
		TimeProvider: mockTime,
		FileOpener:   mockOpener,
	}

	mockOpener.OpenErr = errors.New("open failure")
	_, err = logger.NewLogger(cfgNoFileWriter)
	if err == nil || !strings.Contains(err.Error(), "open failure") {
		t.Errorf("Expected file open error to be 'open failure', got '%v'", err)
	}
}
