package repositories

import (
	"base_lara_go_project/app/models/db"

	"gorm.io/gorm"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) FindByID(id uint) (*db.Service, error) {
	var service db.Service
	err := r.db.First(&service, id).Error
	return &service, err
}

func (r *ServiceRepository) All() ([]db.Service, error) {
	var services []db.Service
	err := r.db.Find(&services).Error
	return services, err
}

// Add more CRUD methods as needed...
