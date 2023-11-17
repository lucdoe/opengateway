package app

import (
	"testing"

	"github.com/lucdoe/capstone_gateway/internal"
	"github.com/lucdoe/capstone_gateway/internal/app"
)

func TestSetupRoutes(t *testing.T) {
	mockRouter := NewMockRouter()
	sampleConfig := &internal.Config{
		Services: map[string]internal.Service{
			"TestService": {
				URL:  "http://localhost",
				PORT: 8080,
				Endpoints: []internal.Endpoint{
					{
						Path:        "/path1",
						HTTPMethod:  "GET",
						AllowedJSON: []string{"field1", "field2"},
					},
					{
						Path:        "/path2",
						HTTPMethod:  "POST",
						AllowedJSON: []string{"field3", "field4"},
					},
				},
			},
		},
	}

	app.SetupRoutes(mockRouter, sampleConfig)

	for _, service := range sampleConfig.Services {
		for _, endpoint := range service.Endpoints {
			expectedRoute := endpoint.HTTPMethod + " " + endpoint.Path
			if _, ok := mockRouter.registeredRoutes[expectedRoute]; !ok {
				t.Errorf("Expected route %s not found", expectedRoute)
			}
		}
	}
}
