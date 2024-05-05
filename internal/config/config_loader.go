package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// JSON or YAML
type ConfigI interface {
}

type ConfigLoaderI interface {
	LoadFromFile(path string) (*TopLevel, error)
}

type Loader struct {
	FileReader FileReader
}

func NewLoader(fileReader FileReader) *Loader {
	return &Loader{FileReader: fileReader}
}

func (l *Loader) LoadConfigFromFile(filePath string) (*TopLevel, error) {
	var config TopLevel

	data, err := l.FileReader.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
