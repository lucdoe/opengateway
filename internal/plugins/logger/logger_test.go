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

package logger

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockFileWriter struct {
	written []byte
}

func (m *MockFileWriter) Write(p []byte) (n int, err error) {
	m.written = append(m.written, p...)
	return len(p), nil
}

type MockFileOpener struct {
	fileWriter *MockFileWriter
}

func (m *MockFileOpener) OpenFile(name string, flag int, perm os.FileMode) (FileWriter, error) {
	return m.fileWriter, nil
}

type MockTimeProvider struct {
	now time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.now
}

func TestNewLogger(t *testing.T) {
	cfg := LoggerConfig{
		FilePath:     "/tmp/test.log",
		FileWriter:   nil,
		ErrOutput:    os.Stderr,
		TimeProvider: RealTime{},
		FileOpener:   DefaultFileOpener{},
	}

	logger := NewLogger(cfg)

	assert.NotNil(t, logger)
	assert.IsType(t, &OSLogger{}, logger)
}

func TestNewLoggerInvalidFilePath(t *testing.T) {
	cfg := LoggerConfig{
		FilePath:     "",
		FileWriter:   nil,
		ErrOutput:    os.Stderr,
		TimeProvider: RealTime{},
		FileOpener:   DefaultFileOpener{},
	}

	assert.Panics(t, func() {
		NewLogger(cfg)
	})
}

func TestOSLoggerInfo(t *testing.T) {
	now := time.Now()
	mockFileWriter := &MockFileWriter{}
	mockTimeProvider := &MockTimeProvider{now: now}
	mockFileOpener := &MockFileOpener{fileWriter: mockFileWriter}

	cfg := LoggerConfig{
		FilePath:     "/tmp/test.log",
		FileWriter:   mockFileWriter,
		ErrOutput:    os.Stderr,
		TimeProvider: mockTimeProvider,
		FileOpener:   mockFileOpener,
	}

	logger := NewLogger(cfg).(*OSLogger)

	logger.Info("Test message", "Test details")

	expectedLogMessage := fmt.Sprintf("%s [%s]: %s %s\n", now.Format("2006-01-02 15:04:05"), "INFO", "Test message", "Test details")
	assert.Equal(t, expectedLogMessage, string(mockFileWriter.written))
}

func TestOSLoggerInfoErrorWritingToFile(t *testing.T) {
	now := time.Now()
	mockFileWriter := &MockFileWriter{}
	mockTimeProvider := &MockTimeProvider{now: now}
	mockFileOpener := &MockFileOpener{fileWriter: mockFileWriter}

	cfg := LoggerConfig{
		FilePath:     "/tmp/test.log",
		FileWriter:   mockFileWriter,
		ErrOutput:    os.Stderr,
		TimeProvider: mockTimeProvider,
		FileOpener:   mockFileOpener,
	}

	logger := NewLogger(cfg).(*OSLogger)

	mockFileWriter.written = []byte{}
	mockFileWriter.Write([]byte(""))

	logger.Info("Test message", "Test details")

	expectedLogMessage := fmt.Sprintf("%s [%s]: %s %s\n", now.Format("2006-01-02 15:04:05"), "INFO", "Test message", "Test details")
	assert.Equal(t, expectedLogMessage, string(mockFileWriter.written))
}

func TestDefaultFileOpenerOpenFile(t *testing.T) {
	fileOpener := DefaultFileOpener{}

	file, err := fileOpener.OpenFile("/tmp/test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	assert.NoError(t, err)
	assert.NotNil(t, file)
}

func TestLoggerConfigsetDefaults(t *testing.T) {
	cfg := LoggerConfig{}

	cfg.setDefaults()

	assert.Equal(t, os.Stderr, cfg.ErrOutput)
	assert.IsType(t, RealTime{}, cfg.TimeProvider)
	assert.IsType(t, DefaultFileOpener{}, cfg.FileOpener)
}
