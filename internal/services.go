package internal

type IService interface {
	GetAllServices() ([]Service, error)
	GetServiceByID(id string) (*Service, error)
	CreateService(*Service) error
	UpdateService(*Service) error
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

func (s *Service) CreateService(srvc *Service) error {
	if err := DB.Create(srvc).Error; err != nil {
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
