package utils

import "gopkg.in/yaml.v3"

type YAMLParsing struct{}

func (YAMLParsing) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}
