package cors_test

import (
	"testing"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cors"
)

func TestValidateOrigin(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Origins: "http://example.com, http://example.org, *",
		Methods: "GET, POST",
		Headers: "Content-Type, Authorization",
	}
	c := cors.NewCors(corsConfig)

	tests := []struct {
		name         string
		origin       string
		expectedPass bool
	}{
		{"Allowed Origin", "http://example.com", true},
		{"Another Allowed Origin", "http://example.org", true},
		{"Disallowed Origin", "http://notallowed.com", false},
		{"Wildcard Origin", "*", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if pass := c.ValidateOrigin(tt.origin); pass != tt.expectedPass {
				t.Errorf("ValidateOrigin(%q) = %v, want %v", tt.origin, pass, tt.expectedPass)
			}
		})
	}
}

func TestValidateMethod(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Origins: "*",
		Methods: "GET, POST",
		Headers: "Content-Type",
	}
	c := cors.NewCors(corsConfig)

	tests := []struct {
		method       string
		expectedPass bool
	}{
		{"GET", true},
		{"POST", true},
		{"DELETE", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			if pass := c.ValidateMethod(tt.method); pass != tt.expectedPass {
				t.Errorf("ValidateMethod(%q) = %v, want %v", tt.method, pass, tt.expectedPass)
			}
		})
	}
}

func TestValidateHeaders(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Origins: "*",
		Methods: "GET, POST",
		Headers: "Content-Type, Authorization",
	}
	c := cors.NewCors(corsConfig)

	tests := []struct {
		headers      string
		expectedPass bool
	}{
		{"Content-Type", true},
		{"Authorization", true},
		{"X-Custom-Header", false},
	}

	for _, tt := range tests {
		t.Run(tt.headers, func(t *testing.T) {
			if pass := c.ValidateHeaders(tt.headers); pass != tt.expectedPass {
				t.Errorf("ValidateHeaders(%q) = %v, want %v", tt.headers, pass, tt.expectedPass)
			}
		})
	}
}
