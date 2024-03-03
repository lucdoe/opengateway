package tests

import (
	"bytes"
	"os"
	"testing"
)

func TestIfReadMeExists(t *testing.T) {
	if _, err := os.Stat("../README.md"); os.IsNotExist(err) {
		t.Errorf("README.md does not exist")
		return
	}

	file, err := os.Open("../README.md")
	if err != nil {
		t.Errorf("README.md could not be opened")
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		t.Errorf("README.md could not be read")
		return
	}

	if !bytes.Contains(buf, []byte("# OpenGateway")) {
		t.Errorf("h1 does not exist in README.md")
		return
	}
}
