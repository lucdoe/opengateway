package internal

type IService interface {
	GetAllServices() ([]Service, error)
}

func (s *Service) GetAllServices() ([]Service, error) {
	var services []Service
	if err := DB.Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}
