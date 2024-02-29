package internal

type IService interface {
	GetAllServices() ([]Service, error)
	GetServiceByID(id string) (*Service, error)
}

func (s *Service) GetAllServices() ([]Service, error) {
	var services []Service
	if err := DB.Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

func (s *Service) GetServiceByID(id string) (*Service, error) {
	var service Service
	if err := DB.First(&service, id).Error; err != nil {
		return nil, err
	}
	return &service, nil
}
