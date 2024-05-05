package internal

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
