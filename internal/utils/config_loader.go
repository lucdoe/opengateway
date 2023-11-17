package utils

import "github.com/lucdoe/capstone_gateway/internal"

type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type YAMLParser interface {
	Unmarshal(in []byte, out interface{}) error
}

type ConfigLoader struct {
	fileReader FileReader
	yamlParser YAMLParser
}

func NewConfigLoader(fr FileReader, yp YAMLParser) *ConfigLoader {
	return &ConfigLoader{
		fileReader: fr,
		yamlParser: yp,
	}
}

func (cl *ConfigLoader) LoadConfig(f string) (*internal.Config, error) {
	data, err := cl.fileReader.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var config internal.Config
	err = cl.yamlParser.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
