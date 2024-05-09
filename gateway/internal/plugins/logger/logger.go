package logger

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

type FileWriter interface {
	Write(p []byte) (n int, err error)
}

type Logger interface {
	Init() error
	Info(msg string, r *http.Request, duration time.Duration)
	Error(msg string)
}

type OSLogger struct {
	filePath string
	file     FileWriter
}

func NewLogger(filePath string, file FileWriter) Logger {
	return &OSLogger{
		filePath: filePath,
		file:     file,
	}
}

func (l *OSLogger) Init() error {
	if l.file == nil {
		var err error
		l.file, err = os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
	}
	return nil
}

func (l *OSLogger) Info(msg string, r *http.Request, duration time.Duration) {
	if r != nil {
		msg += fmt.Sprintf(" | Request: %s %s from %s in %dms", r.Method, r.URL.Path, r.RemoteAddr, duration.Milliseconds())
	}
	l.log("INFO", msg)
}

func (l *OSLogger) Error(msg string) {
	l.log("ERROR", msg)
}

func (l *OSLogger) log(level, msg string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("%s [%s]: %s\n", timestamp, level, msg)

	if l.file != nil {
		_, err := l.file.Write([]byte(logMessage))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to log file: %v\n", err)
		}
	}
}
