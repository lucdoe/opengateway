package config

import (
	"fmt"
	"os"
)

type FileReader interface {
	ReadFile(filePath string) ([]byte, error)
}

type OSFileReader struct{}

func (fr *OSFileReader) ReadFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	return data, nil
}
