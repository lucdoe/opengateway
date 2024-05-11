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
	Info(msg string, r *http.Request)
	Middleware(next http.Handler) http.Handler
}

type OSLogger struct {
	filePath string
	file     FileWriter
}

func NewLogger(filePath string, file FileWriter) (Logger, error) {
	if file == nil {
		var err error
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	return &OSLogger{
		filePath: filePath,
		file:     file,
	}, nil
}

func (l *OSLogger) Info(msg string, r *http.Request) {
	if r != nil {
		msg += fmt.Sprintf(" %s: %s %s from %s", msg, r.Method, r.URL.Path, r.RemoteAddr)
	}
	l.log("INFO", msg)
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

func (l *OSLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.Info("Request", r)
		next.ServeHTTP(w, r)
	})
}
