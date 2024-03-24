package config

type Reader interface {
	ReadFile(name string) ([]byte, error)
}

type concreteFileReader struct {
	reader FileReader
}

type FileReader interface {
	ReadFile(name string) ([]byte, error)
}

func NewFileReader(reader FileReader) Reader {
	return &concreteFileReader{reader: reader}
}

func (r *concreteFileReader) ReadFile(name string) ([]byte, error) {
	return r.reader.ReadFile(name)
}
