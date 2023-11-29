package utils

import (
	"errors"
	"testing"

	"github.com/lucdoe/capstone_gateway/internal"
)

type MockFileReader struct {
	FileContent []byte
	ReadErr     error
}

func (mfr *MockFileReader) ReadFile(filename string) ([]byte, error) {
	return mfr.FileContent, mfr.ReadErr
}

type MockYAMLParser struct {
	UnmarshalErr error
}

func (myp *MockYAMLParser) Unmarshal(in []byte, out interface{}) error {
	if myp.UnmarshalErr != nil {
		return myp.UnmarshalErr
	}
	*out.(*internal.Config) = internal.Config{
		Services: map[string]internal.Service{
			"capstone_gateway": {
				PORT:      9876,
				URL:       "http://localhost",
				SecretKey: "verySecretKey",
				Endpoints: []internal.Endpoint{
					{
						Name:       "GETTest",
						HTTPMethod: "GET",
						Path:       "/test",
						AllowedJSON: []string{
							"testBody",
						},
					},
				},
			},
		},
	}
	return nil
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		fileContent []byte
		readErr     error
		parseErr    error
		expectErr   bool
	}{
		{"Successful Load", []byte("valid yaml content"), nil, nil, false},
		{"Read File Error", []byte(""), errors.New("read error"), nil, true},
		{"Parse Error", []byte("invalid yaml content"), nil, errors.New("parse error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mfr := &MockFileReader{FileContent: tt.fileContent, ReadErr: tt.readErr}
			myp := &MockYAMLParser{UnmarshalErr: tt.parseErr}

			cl := internal.NewConfigLoader(mfr, myp)

			_, err := cl.LoadConfig("test_config.yaml")

			if (err != nil) != tt.expectErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.expectErr)
			}
		})
	}
}
