package middlewares

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const logFilePathTemplate = "logs/%s.log"

func setupLogger() *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	f, err := os.OpenFile(getLogFileName(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		l.Out = f
	} else {
		l.Info("Failed to log to file, using default stderr")
	}

	return l
}

func getLogFileName() string {
	return fmt.Sprintf(logFilePathTemplate, time.Now().Format(time.DateOnly))
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

func logRequestDetails(c *gin.Context, l *logrus.Logger) {
	start := time.Now()
	c.Next()
	latency := time.Since(start)

	log := l.WithFields(logrus.Fields{
		"http.method":               c.Request.Method,
		"url.original":              c.Request.URL.Path,
		"http.response.status_code": c.Writer.Status(),
		"event.duration":            int(latency.Milliseconds()),
		"client.ip":                 maskIP(c.ClientIP()),
		"user_agent.original":       c.Request.UserAgent(),
		"@timestamp":                start.Format(time.RFC3339),
		"event.kind":                "info",
		"event.outcome":             "success",
		"http.request.referrer":     c.Request.Referer(),
	})

	if len(c.Errors) > 0 {
		log.Error("Request error")
	} else {
		log.Info("Request processed")
	}
}

func LogRequest(c *gin.Context) {
	l := setupLogger()
	logRequestDetails(c, l)
}
