package middlewares

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const logFilePathTemplate = "logs/%s.log"

type LogLevel int

const (
	INFO LogLevel = iota
	ERROR
)

type Logger struct {
	Out *os.File
}

func CustomLogger() *Logger {
	l := &Logger{}
	f, err := os.OpenFile(getLogFileName(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		l.Out = f
	} else {
		l.Out = os.Stderr
		fmt.Println("Failed to log to file, using default stderr")
	}
	return l
}

func (l *Logger) log(level LogLevel, fields map[string]interface{}) {
	fields["@timestamp"] = time.Now().Format(time.RFC3339)
	var levelStr string
	switch level {
	case INFO:
		levelStr = "info"
	case ERROR:
		levelStr = "error"
	}
	fields["event.kind"] = levelStr
	logData, _ := json.Marshal(fields)
	l.Out.Write(logData)
	l.Out.WriteString("\n")
}

func (l *Logger) Info(fields map[string]interface{}) {
	l.log(INFO, fields)
}

func (l *Logger) Error(fields map[string]interface{}) {
	l.log(ERROR, fields)
}

func setupLogger() *Logger {
	return CustomLogger()
}

func getLogFileName() string {
	return fmt.Sprintf(logFilePathTemplate, time.Now().Format("2006-01-02"))
}

func maskIP(ip string) string {
	IPv6 := strings.Count(ip, ":") >= 2
	IPv4 := strings.Count(ip, ".") >= 3

	if IPv6 {
		i := strings.LastIndex(ip, ":")
		return ip[:i] + ":***"
	} else if IPv4 {
		i := strings.LastIndex(ip, ".")
		return ip[:i] + ".***"
	}
	return ip
}

func logRequestDetails(c *gin.Context, l *Logger) {
	start := time.Now()
	c.Next()
	latency := time.Since(start)

	logFields := map[string]interface{}{
		"http.method":               c.Request.Method,
		"url.original":              c.Request.URL.Path,
		"http.response.status_code": c.Writer.Status(),
		"event.duration":            int(latency.Milliseconds()),
		"client.ip":                 maskIP(c.ClientIP()),
		"user_agent.original":       c.Request.UserAgent(),
		"http.request.referrer":     c.Request.Referer(),
		"event.outcome":             "success",
	}

	if len(c.Errors) > 0 {
		logFields["event.outcome"] = "failure"
		l.Error(logFields)
	} else {
		l.Info(logFields)
	}
}

func LogRequest(c *gin.Context) {
	l := setupLogger()
	logRequestDetails(c, l)
}
