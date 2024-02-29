package databases

import (
	"gorm.io/gorm"
)

// ServiceRepository interface abstracts the
// data access operations for services.
type ServiceRepository interface {
	GetAll() ([]Service, error)
	GetByID(id string) (*Service, error)
	Create(service *Service) error
	Update(service *Service) error
}

type GormServiceRepository struct {
	DB *gorm.DB
}

func NewGormServiceRepository(db *gorm.DB) *GormServiceRepository {
	return &GormServiceRepository{DB: db}
}

func (r *GormServiceRepository) GetAll() ([]Service, error) {
	var services []Service

	if err := r.DB.Preload("Protocol").Preload("Endpoints").Preload("Middlewares").Preload("Tags").Find(&services).Error; err != nil {
		return nil, err
	}

	return services, nil
}

func (r *GormServiceRepository) GetByID(id string) (*Service, error) {
	var service Service

	if err := r.DB.Preload("Protocol").Preload("Endpoints").Preload("Middlewares").Preload("Tags").First(&service, id).Error; err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *GormServiceRepository) Create(service *Service) error {
	return r.DB.Create(service).Error
}

func (r *GormServiceRepository) Update(service *Service) error {
	return r.DB.Save(service).Error
}
