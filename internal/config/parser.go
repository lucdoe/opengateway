package config

type Parser interface {
	Unmarshal(data []byte) (map[string]string, error)
}

type concreteYAMLParser struct {
	parser YAMLParser
}

type YAMLParser interface {
	Unmarshal(YAMLData []byte) (map[string]string, error)
}

func NewYAMLParser(parser YAMLParser) Parser {
	return &concreteYAMLParser{parser: parser}
}

func (p *concreteYAMLParser) Unmarshal(data []byte) (map[string]string, error) {
	return p.parser.Unmarshal(data)
}
