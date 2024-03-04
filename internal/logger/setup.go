package logger

import (
	"log"
	"os"
	"time"
)

type LoggerSetup interface {
	InitialiseCustomLogger()
}

type StandardLogger struct{}

func (sl *StandardLogger) InitialiseCustomLogger() {
	timeAndDate := time.Now().Format("2006-01-02") // Use reference time format

	logFile, err := os.OpenFile("logs/"+timeAndDate+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file: ", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func (sl *StandardLogger) Log(message string) {
	log.Println(message)
}
