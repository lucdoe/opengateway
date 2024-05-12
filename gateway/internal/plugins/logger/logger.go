package logger

import (
	"fmt"
	"os"
	"time"
)

type FileWriter interface {
	Write(p []byte) (n int, err error)
}

type Logger interface {
	Info(msg string, details string)
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

func (l *OSLogger) Info(msg string, details string) {
	logMessage := fmt.Sprintf("%s [%s]: %s %s\n", time.Now().Format("2006-01-02 15:04:05"), "INFO", msg, details)
	if l.file != nil {
		_, err := l.file.Write([]byte(logMessage))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to log file: %v\n", err)
		}
	}
}
