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

type OSLogger struct {
	filePath     string
	file         FileWriter
	errOutput    io.Writer
	timeProvider TimeProvider
	fileOpener   FileOpener
}

func NewLogger(filePath string, file FileWriter, errOutput io.Writer, timeProvider TimeProvider, fileOpener FileOpener) (Logger, error) {
	if file == nil {
		if fileOpener == nil {
			fileOpener = DefaultFileOpener{}
		}
		var err error
		file, err = fileOpener.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}
	if errOutput == nil {
		errOutput = os.Stderr
	}
	if timeProvider == nil {
		timeProvider = RealTime{}
	}
	return &OSLogger{
		filePath:     filePath,
		file:         file,
		errOutput:    errOutput,
		timeProvider: timeProvider,
		fileOpener:   fileOpener,
	}, nil
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
