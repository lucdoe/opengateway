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

package cors

import (
	"strings"
)

type CORS interface {
	ValidateOrigin(origin string) bool
	ValidateMethod(method string) bool
	ValidateHeaders(headers string) bool
	GetAllowedMethods() string
	GetAllowedHeaders() string
}

type CORSConfig struct {
	Origins string
	Methods string
	Headers string
}

func NewCors(config CORSConfig) *CORSConfig {
	return &CORSConfig{Origins: config.Origins, Methods: config.Methods, Headers: config.Headers}
}

func (c *CORSConfig) ValidateOrigin(origin string) bool {
	if c.Origins == "*" {
		return true
	}
	for _, o := range strings.Split(c.Origins, ", ") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func (c *CORSConfig) ValidateMethod(method string) bool {
	for _, m := range strings.Split(c.Methods, ", ") {
		if strings.TrimSpace(m) == method {
			return true
		}
	}
	return false
}

func (c *CORSConfig) ValidateHeaders(requestedHeaders string) bool {
	if requestedHeaders == "" {
		return true
	}
	allowedHeaders := strings.Split(c.Headers, ", ")
	for _, header := range strings.Split(requestedHeaders, ",") {
		headerFound := false
		for _, allowedHeader := range allowedHeaders {
			if strings.TrimSpace(allowedHeader) == strings.TrimSpace(header) {
				headerFound = true
				break
			}
		}
		if !headerFound {
			return false
		}
	}
	return true
}

func (c *CORSConfig) GetAllowedMethods() string {
	return c.Methods
}

func (c *CORSConfig) GetAllowedHeaders() string {
	return c.Headers
}
