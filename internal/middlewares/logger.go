package middlewares

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	logFilePathTemplate        = "logs/%s.log"
	INFO                LogLVL = iota
	ERROR
)

type LogLVL int
type Logger struct {
	Out *os.File
}

func getLogFileName() string {
	return fmt.Sprintf(logFilePathTemplate, time.Now().Format("2006-01-02"))
}

func (l *Logger) log(lvl LogLVL, fields map[string]interface{}) {
	var lvlStr string

	fields["@timestamp"] = time.Now().Format(time.RFC3339)

	switch lvl {
	case INFO:
		lvlStr = "info"
	case ERROR:
		lvlStr = "error"
	}

	fields["event.kind"] = lvlStr
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

func customLogger() *Logger {
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

func logRequestDetails(c *gin.Context, l *Logger) {
	start := time.Now()
	c.Next()
	latency := time.Since(start)

	logFields := map[string]interface{}{
		"http.method":               c.Request.Method,
		"url.original":              c.Request.URL.Path,
		"http.response.status_code": c.Writer.Status(),
		"event.duration":            int(latency.Milliseconds()),
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
	l := customLogger()
	logRequestDetails(c, l)
}
