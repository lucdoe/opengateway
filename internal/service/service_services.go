package service

import (
	"fmt"
	"time"

	"github.com/lucdoe/gateway_admin_api/internal/databases"
)

type IService interface {
	GetAll() ([]ServiceResponse, error)
	GetByID(id string) (*ServiceResponse, error)
	Create(*databases.Service) error
	Update(*databases.Service) error
}

type ServiceResponse struct {
	ID          uint                   `json:"id"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Name        string                 `json:"name"`
	Host        string                 `json:"host"`
	Port        string                 `json:"port"`
	Enabled     bool                   `json:"enabled"`
	Protocol    databases.Protocol     `json:"protocol"`    // Assuming ProtocolResponse is defined similarly
	Endpoints   []databases.Endpoint   `json:"endpoints"`   // Define EndpointResponse similarly
	Middlewares []databases.Middleware `json:"middlewares"` // Define MiddlewareResponse similarly
	Tags        []databases.Tag        `json:"tags"`        // Define TagResponse similarly
	Links       []Link                 `json:"links"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type ServiceImplementation struct {
	Repo databases.ServiceRepository
}

func NewServiceImplementation(repo databases.ServiceRepository) *ServiceImplementation {
	return &ServiceImplementation{Repo: repo}
}

func (si *ServiceImplementation) GetAll() ([]ServiceResponse, error) {
	services, err := si.Repo.GetAll()
	if err != nil {
		return nil, err
	}

	var serviceResponses []ServiceResponse
	var sendLinks bool = true
	for _, service := range services {
		response := mapServiceToServiceResponse(service, sendLinks)
		serviceResponses = append(serviceResponses, response)
	}

	return serviceResponses, nil
}

func (si *ServiceImplementation) GetByID(id string) (*ServiceResponse, error) {
	service, err := si.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	var sendNoLinks bool = false
	response := mapServiceToServiceResponse(*service, sendNoLinks)

	return &response, nil
}

func (si *ServiceImplementation) Create(srvc *databases.Service) error {
	if err := si.Repo.Create(srvc); err != nil {
		return err
	}
	return nil
}

func (si *ServiceImplementation) Update(srvc *databases.Service) error {
	if err := si.Repo.Update(srvc); err != nil {
		return err
	}
	return nil
}

func mapServiceToServiceResponse(service databases.Service, link bool) ServiceResponse {
	if link {
		return ServiceResponse{
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
	} else {
		return ServiceResponse{
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
	}
}
