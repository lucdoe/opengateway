package utils

import "os"

type OSFileReader struct{}

func (OSFileReader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
