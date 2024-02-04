package app

import (
	"github.com/lucdoe/capstone_gateway/internal"
)

func LoadConfig(configPath string) (*internal.Config, error) {
	fileReader := internal.OSFileReader{}
	yamlParser := internal.YAMLParsing{}
	configLoader := internal.NewConfigLoader(fileReader, yamlParser)

	return configLoader.LoadConfig(configPath)
}
