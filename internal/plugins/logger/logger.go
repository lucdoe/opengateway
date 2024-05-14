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
	"io"
	"os"
	"time"
)

type FileWriter interface {
	Write(p []byte) (n int, err error)
}

type FileOpener interface {
	OpenFile(name string, flag int, perm os.FileMode) (FileWriter, error)
}

type Logger interface {
	Info(msg string, details string)
}

type TimeProvider interface {
	Now() time.Time
}

type LoggerConfig struct {
	FilePath     string
	FileWriter   FileWriter
	ErrOutput    io.Writer
	TimeProvider TimeProvider
	FileOpener   FileOpener
}

type OSLogger struct {
	filePath     string
	file         FileWriter
	errOutput    io.Writer
	timeProvider TimeProvider
}

func NewLogger(cfg LoggerConfig) Logger {
	cfg.setDefaults()

	var err error
	cfg.FileWriter, err = cfg.FileOpener.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)

	}

	return &OSLogger{
		filePath:     cfg.FilePath,
		file:         cfg.FileWriter,
		errOutput:    cfg.ErrOutput,
		timeProvider: cfg.TimeProvider,
	}
}

func (l *OSLogger) Info(msg string, details string) {
	logMessage := fmt.Sprintf("%s [%s]: %s %s\n", l.timeProvider.Now().Format("2006-01-02 15:04:05"), "INFO", msg, details)
	if l.file != nil {
		_, err := l.file.Write([]byte(logMessage))
		if err != nil {
			fmt.Fprintf(l.errOutput, "Error writing to log file: %v\n", err)
		}
	}
}

type RealTime struct{}

func (r RealTime) Now() time.Time {
	return time.Now()
}

type DefaultFileOpener struct{}

func (d DefaultFileOpener) OpenFile(name string, flag int, perm os.FileMode) (FileWriter, error) {
	return os.OpenFile(name, flag, perm)
}

func (cfg *LoggerConfig) setDefaults() {
	if cfg.ErrOutput == nil {
		cfg.ErrOutput = os.Stderr
	}
	if cfg.TimeProvider == nil {
		cfg.TimeProvider = RealTime{}
	}
	if cfg.FileOpener == nil {
		cfg.FileOpener = DefaultFileOpener{}
	}
}
