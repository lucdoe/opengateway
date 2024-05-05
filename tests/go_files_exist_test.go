package tests

import (
	"bytes"
	"os"
	"testing"
)

func TestMainFunctionExists(t *testing.T) {
	file, err := os.Open("../cmd/opengateway/main.go")
	if err != nil {
		t.Error("main.go could not be opened")
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		t.Error("main.go could not be read")
	}

	if !bytes.Contains(buf, []byte("func main()")) {
		t.Error("main function does not exist in main.go")
	}
}

func TestGoModExists(t *testing.T) {
	if _, err := os.Stat("../go.mod"); os.IsNotExist(err) {
		t.Error("go.mod does not exist")
	}

	file, err := os.Open("../go.mod")
	if err != nil {
		t.Error("go.mod could not be opened")
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		t.Error("go.mod could not be read")
	}

	if !bytes.Contains(buf, []byte("module github.com/lucdoe/opengateway")) {
		t.Error("module name in go.mod is incorrect")
	}

	if !bytes.Contains(buf, []byte("go 1.21.1")) {
		t.Error("go version in go.mod is incorrect")
	}
}
