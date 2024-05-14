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

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type QueryParam struct {
	Key   string `yaml:"Key"`
	Value string `yaml:"Value"`
}

type AuthConfig struct {
	ApplyAuth bool   `yaml:"ApplyAuth"`
	Method    string `yaml:"Method"`
	Algorithm string `yaml:"Algorithm"`
	Scope     string `yaml:"Scope"`
	SecretKey string `yaml:"SecretKey"`
}

type Endpoint struct {
	Name        string                 `yaml:"Name"`
	HTTPMethod  string                 `yaml:"HTTPMethod"`
	Path        string                 `yaml:"Path"`
	QueryParams []QueryParam           `yaml:"QueryParams"`
	Auth        AuthConfig             `yaml:"Auth"`
	Body        map[string]interface{} `yaml:"Body"`
	Plugins     []string               `yaml:"Plugins"`
}

type Service struct {
	URL       string     `yaml:"URL"`
	Protocol  string     `yaml:"Protocol"`
	Endpoints []Endpoint `yaml:"Endpoints"`
	Plugins   []string   `yaml:"Plugins"`
}

type Config struct {
	Services map[string]Service `yaml:"Services"`
}

type Parser interface {
	ReadFile() ([]byte, error)
	Unmarshal(in []byte, out interface{}) error
	Parse() (*Config, error)
}

type YAMLParser struct {
	path string
}

func NewParser(path string) Parser {
	return &YAMLParser{
		path: path,
	}
}

func (p *YAMLParser) ReadFile() ([]byte, error) {
	return os.ReadFile(p.path)
}

func (p *YAMLParser) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

func (p *YAMLParser) Parse() (*Config, error) {
	data, err := p.ReadFile()
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = p.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
