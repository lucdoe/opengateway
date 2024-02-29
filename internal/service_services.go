package internal

import (
	"fmt"
	"time"
)

type IService interface {
	GetAllServices() ([]ServiceResponse, error)
	GetServiceByID(id string) (*ServiceResponse, error)
	CreateService(*Service) error
	UpdateService(*Service) error
}

type ServiceResponse struct {
	ID          uint         `json:"id"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	Name        string       `json:"name"`
	Host        string       `json:"host"`
	Port        string       `json:"port"`
	Enabled     bool         `json:"enabled"`
	Protocol    Protocol     `json:"protocol"`    // Assuming ProtocolResponse is defined similarly
	Endpoints   []Endpoint   `json:"endpoints"`   // Define EndpointResponse similarly
	Middlewares []Middleware `json:"middlewares"` // Define MiddlewareResponse similarly
	Tags        []Tag        `json:"tags"`        // Define TagResponse similarly
	Links       []Link       `json:"links"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func (s *Service) GetAllServices() ([]ServiceResponse, error) {
	var services []Service
	var serviceResponses []ServiceResponse

	if err := DB.Preload("Protocol").Preload("Endpoints").Preload("Middlewares").Preload("Tags").Find(&services).Error; err != nil {
		return nil, err
	}

	for _, service := range services {
		response := ServiceResponse{
			ID:          service.ID,
			CreatedAt:   service.CreatedAt,
			UpdatedAt:   service.UpdatedAt,
			Name:        service.Name,
			Host:        service.Host,
			Port:        service.Port,
			Enabled:     service.Enabled,
			Protocol:    service.Protocol,
			Endpoints:   service.Endpoints,
			Middlewares: service.Middlewares,
			Tags:        service.Tags,
			Links: []Link{
				{
					Rel:  "self",
					Href: fmt.Sprintf("%s/%d", ServicesPath, service.ID),
				},
			},
		}
		serviceResponses = append(serviceResponses, response)
	}

	return serviceResponses, nil
}

func (s *Service) GetServiceByID(id string) (*ServiceResponse, error) {
	var service Service
	var serviceResponse ServiceResponse

	if err := DB.First(&service, id).Error; err != nil {
		return nil, err
	}

	serviceResponse = ServiceResponse{
		ID:          service.ID,
		CreatedAt:   service.CreatedAt,
		UpdatedAt:   service.UpdatedAt,
		Name:        service.Name,
		Host:        service.Host,
		Port:        service.Port,
		Enabled:     service.Enabled,
		Protocol:    service.Protocol,
		Endpoints:   service.Endpoints,
		Middlewares: service.Middlewares,
		Tags:        service.Tags,
	}

	return &serviceResponse, nil
}

func (s *Service) CreateService(srvc *Service) error {
	if err := DB.Preload("Protocol").Create(srvc).Error; err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateService(srvc *Service) error {
	if err := DB.Save(srvc).Error; err != nil {
		return err
	}
	return nil
}
