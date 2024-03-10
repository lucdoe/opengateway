package config

type TopLevel struct {
	Title             string       `yaml:"title" json:"title"`
	Contact           Contact      `yaml:"contact"  json:"contact"`
	Version           string       `yaml:"version" json:"version"`
	GlobalMiddlewares []Middleware `yaml:"globalMiddlewares" json:"globalMiddlewares"`
	Services          []Service    `yaml:"services" json:"services"`
}

type Contact struct {
	Name  string `yaml:"name" json:"name"`
	URL   string `yaml:"url" json:"url"`
	Email string `yaml:"email" json:"email"`
}

type Middleware struct {
	Name        string   `yaml:"name" json:"name"`
	Middleware  string   `yaml:"middleware" json:"middleware"`
	Description string   `yaml:"description" json:"description"`
	Config      *Config  `yaml:"config,omitempty" json:"config,omitempty"`
	Methods     []string `yaml:"methods" json:"methods"`
}

type Config struct {
	Methods                 []string `yaml:"methods" json:"methods"`
	Headers                 []string `yaml:"headers" json:"headers"`
	Origins                 []string `yaml:"origins" json:"origins"`
	Credentials             *string  `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	MaxAge                  *int64   `yaml:"maxAge,omitempty" json:"maxAge,omitempty"`
	XSSProtection           *string  `yaml:"xssProtection,omitempty" json:"xssProtection,omitempty"`
	ContentTypeOptions      *string  `yaml:"contentTypeOptions,omitempty" json:"contentTypeOptions,omitempty"`
	FrameOptions            *string  `yaml:"frameOptions,omitempty" json:"frameOptions,omitempty"`
	ReferrerPolicy          *string  `yaml:"referrerPolicy,omitempty" json:"referrerPolicy,omitempty"`
	FeaturePolicy           *string  `yaml:"featurePolicy,omitempty" json:"featurePolicy,omitempty"`
	ContentSecurityPolicy   *string  `yaml:"contentSecurityPolicy,omitempty" json:"contentSecurityPolicy,omitempty"`
	PermissionsPolicy       *string  `yaml:"permissionsPolicy,omitempty" json:"permissionsPolicy,omitempty"`
	StrictTransportSecurity *string  `yaml:"strictTransportSecurity,omitempty" json:"strictTransportSecurity,omitempty"`
	Max                     *int64   `yaml:"max,omitempty" json:"max,omitempty"`
	Duration                *string  `yaml:"duration,omitempty" json:"duration,omitempty"`
	Limit                   *string  `yaml:"limit,omitempty" json:"limit,omitempty"`
	Secret                  *string  `yaml:"secret,omitempty" json:"secret,omitempty"`
}

type Service struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description" json:"description"`
	Contact     ServiceContact `yaml:"contact" json:"contact"`
	Protocol    string         `yaml:"protocol" json:"protocol"`
	Host        string         `yaml:"host" json:"host"`
	BasePath    string         `yaml:"basePath" json:"basePath"`
	Port        int64          `yaml:"port" json:"port"`
	URL         string         `yaml:"url" json:"url"`
	Endpoints   []Endpoint     `yaml:"endpoints" json:"endpoints"`
}

type ServiceContact struct {
	Name  string `yaml:"name" json:"name"`
	Email string `yaml:"email" json:"email"`
	Note  string `yaml:"note" json:"note"`
}

type Endpoint struct {
	Path        string       `yaml:"path" json:"path"`
	Methods     []string     `yaml:"methods" json:"methods"`
	Description string       `yaml:"description" json:"description"`
	Middlewares []Middleware `yaml:"middlewares" json:"middlewares"`
}
