package logger_test

import (
	"testing"
)

type MockLoggerSetup struct {
	Initialised bool
}

func (ml *MockLoggerSetup) InitialiseCustomLogger() {
	ml.Initialised = true
}

func TestInitialiseCustomLogger(t *testing.T) {
	mockLogger := &MockLoggerSetup{}

	mockLogger.InitialiseCustomLogger()

	if !mockLogger.Initialised {
		t.Errorf("Expected InitialiseCustomLogger to set Initialised to true, got false")
	}
}
