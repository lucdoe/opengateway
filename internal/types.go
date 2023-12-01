package internal

import "github.com/gin-gonic/gin"

type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type YAMLParser interface {
	Unmarshal(in []byte, out interface{}) error
}

type RouterInterface interface {
	GET(path string, handlers ...gin.HandlerFunc)
	POST(path string, handlers ...gin.HandlerFunc)
	PUT(path string, handlers ...gin.HandlerFunc)
	PATCH(path string, handlers ...gin.HandlerFunc)
	Use(middlewares ...gin.HandlerFunc)
	Run(addr ...string) error
}

type CORSConfig struct {
	Apply   bool     `yaml:"Apply"`
	Origins []string `yaml:"Origins"`
	Headers []string `yaml:"Headers"`
}

type CompressionConfig struct {
	Apply bool `yaml:"Apply"`
}

type QueryParam struct {
	Key  string `yaml:"Key"`
	Type string `yaml:"Type"`
}

type AuthScope struct {
	ApplyScope bool     `yaml:"ApplyScope"`
	Names      []string `yaml:"Names"`
}

type AuthConfig struct {
	ApplyAuth bool      `yaml:"ApplyAuth"`
	Method    string    `yaml:"Method"`
	Algorithm string    `yaml:"Algorithm"`
	Scope     AuthScope `yaml:"Scope"`
}

type BodyField struct {
	ApplyValidation bool                   `yaml:"ApplyValidation"`
	Type            string                 `yaml:"Type"`
	Fields          map[string]interface{} `yaml:"Fields"`
}

type ResilianceConfig struct {
	ApplyRateLimit    bool `yaml:"ApplyRateLimit"`
	RequestsPerMinute int  `yaml:"RequestsPerMinute"`
}

type Endpoint struct {
	Name        string           `yaml:"Name"`
	HTTPMethod  string           `yaml:"HTTPMethod"`
	Path        string           `yaml:"Path"`
	QueryParams []QueryParam     `yaml:"QueryParams"`
	Auth        AuthConfig       `yaml:"Auth"`
	Body        BodyField        `yaml:"Body"`
	Resiliance  ResilianceConfig `yaml:"Resiliance"`
}

type Service struct {
	PORT      int        `yaml:"PORT"`
	URL       string     `yaml:"URL"`
	SecretKey string     `yaml:"SecretKey"`
	Protocol  string     `yaml:"Protocol"`
	Endpoints []Endpoint `yaml:"Endpoints"`
}

type Config struct {
	CORS        CORSConfig         `yaml:"CORS"`
	Compression CompressionConfig  `yaml:"Compression"`
	Services    map[string]Service `yaml:"Services"`
}

type App struct {
	Router RouterInterface
}

type ConfigLoader struct {
	fileReader FileReader
	yamlParser YAMLParser
}
