package internal

func NewConfigLoader(fr FileReader, yp YAMLParser) *ConfigLoader {
	return &ConfigLoader{
		fileReader: fr,
		yamlParser: yp,
	}
}

func (cl *ConfigLoader) LoadConfig(f string) (*Config, error) {
	data, err := cl.fileReader.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var config Config
	err = cl.yamlParser.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
