// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cors_test

import (
	"testing"

	"github.com/lucdoe/opengateway/internal/plugins/cors"
)

const (
	GETAndPOST                  = "GET, POST"
	GETandPOSTandPUT            = "GET, POST, PUT"
	contentTypeAuthzXReqHeaders = "Content-Type, Authorization, X-Requested-With"
	contentTypeAuthzHeaders     = "Content-Type, Authorization"
)

func TestValidateOrigin(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Origins: "http://example.com, http://example.org, *",
		Methods: GETAndPOST,
		Headers: contentTypeAuthzHeaders,
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
		Methods: GETAndPOST,
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
		{"PUT", false},
		{"OPTIONS", false},
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
		Methods: GETAndPOST,
		Headers: contentTypeAuthzHeaders,
	}
	c := cors.NewCors(corsConfig)

	tests := []struct {
		headers      string
		expectedPass bool
	}{
		{"Content-Type", true},
		{"Authorization", true},
		{"X-Custom-Header", false},
		{contentTypeAuthzHeaders, true},
		{"Content-Type, X-Custom-Header", false},
		{"", true}, // No headers requested, should pass
	}

	for _, tt := range tests {
		t.Run(tt.headers, func(t *testing.T) {
			if pass := c.ValidateHeaders(tt.headers); pass != tt.expectedPass {
				t.Errorf("ValidateHeaders(%q) = %v, want %v", tt.headers, pass, tt.expectedPass)
			}
		})
	}
}

func TestGetAllowedMethods(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Methods: GETandPOSTandPUT,
	}
	c := cors.NewCors(corsConfig)

	expected := GETandPOSTandPUT
	if methods := c.GetAllowedMethods(); methods != expected {
		t.Errorf("GetAllowedMethods() = %v, want %v", methods, expected)
	}
}

func TestGetAllowedHeaders(t *testing.T) {
	corsConfig := cors.CORSConfig{
		Headers: contentTypeAuthzXReqHeaders,
	}
	c := cors.NewCors(corsConfig)

	expected := contentTypeAuthzXReqHeaders
	if headers := c.GetAllowedHeaders(); headers != expected {
		t.Errorf("GetAllowedHeaders() = %v, want %v", headers, expected)
	}
}
