package utils_test

import (
	"testing"

	"github.com/lucdoe/opengateway/internal/utils"
)

func TestConstructURL(t *testing.T) {
	constructor := utils.GatewayURLConstructor()

	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := constructor.ConstructURL(tc.baseURL, tc.path)

			if (err != nil) != tc.wantErr {
				t.Errorf("ConstructURL() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && got.String() != tc.want {
				t.Errorf("ConstructURL() got = %v, want %v", got, tc.want)
			}
		})
	}
}
