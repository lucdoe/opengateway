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

type Endpoint struct {
	Name        string   `yaml:"Name"`
	HTTPMethod  string   `yaml:"HTTPMethod"`
	Path        string   `yaml:"Path"`
	AllowedJSON []string `yaml:"AllowedJSON"`
}

type Service struct {
	PORT      int        `yaml:"PORT"`
	URL       string     `yaml:"URL"`
	SecretKey string     `yaml:"SecretKey"`
	Endpoints []Endpoint `yaml:"Endpoints"`
}

type Config struct {
	Services map[string]Service `yaml:"Services"`
}

type App struct {
	Router RouterInterface
}

type ConfigLoader struct {
	fileReader FileReader
	yamlParser YAMLParser
}
