package internal

import (
	"os"

	"gopkg.in/yaml.v3"
)

type OSFileReader struct{}
type YAMLParsing struct{}

func (OSFileReader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (YAMLParsing) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}
