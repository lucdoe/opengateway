package utils__test

import (
	"testing"

	"github.com/lucdoe/opengateway/internal/utils"
)

func TestConstructURL(t *testing.T) {
	// Define test cases
	tests := []struct {
		name    string
		baseURL string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid HTTP URL",
			baseURL: "http://example.com",
			path:    "/test",
			want:    "http://example.com/test",
			wantErr: false,
		},
		{
			name:    "Valid HTTPS URL with trailing slash",
			baseURL: "https://example.com/",
			path:    "test",
			want:    "https://example.com/test",
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			baseURL: "http://a b.com/",
			path:    "test",
			wantErr: true,
		},
	}

	for _, testcase := range tests {

		t.Run(testcase.name, func(t *testing.T) {

			got, err := utils.ConstructURL(testcase.baseURL, testcase.path)

			if (err != nil) != testcase.wantErr {
				t.Errorf("ConstructURL() error = %v, wantErr %v", err, testcase.wantErr)
				return
			}

			if !testcase.wantErr && got.String() != testcase.want {
				t.Errorf("ConstructURL() got = %v, want %v", got, testcase.want)
			}

		})
	}
}
