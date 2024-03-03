package tests

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestIfGitSetupExists(t *testing.T) {
	if _, err := os.Stat("../.git"); os.IsNotExist(err) {
		t.Errorf(".git does not exist")
	}
}

func TestIfGitIgnoreExists(t *testing.T) {
	if _, err := os.Stat("../.gitignore"); os.IsNotExist(err) {
		t.Errorf(".gitignore does not exist")
		return
	}

	file, err := os.Open("../.gitignore")
	if err != nil {
		t.Errorf(".gitignore could not be opened")
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		t.Errorf(".gitignore could not be read")
		return
	}

	entries := []string{
		"*.exe",
		"*.dll",
		"*.so",
		"*.dylib",
		"*.test",
		"*.outgo",
		".work",
		"bin/",
		"logs/",
		"*.log",
		".env",
		"*.env.*",
	}

	for _, entry := range entries {
		if !bytes.Contains(buf, []byte(entry)) {
			t.Errorf("%s does not exist in .gitignore", entry)
		}
	}
}
