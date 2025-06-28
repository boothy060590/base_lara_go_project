package repositories

import (
	"base_lara_go_project/app/models/db"

	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) FindByID(id uint) (*db.Permission, error) {
	var permission db.Permission
	err := r.db.First(&permission, id).Error
	return &permission, err
}

func (r *PermissionRepository) All() ([]db.Permission, error) {
	var permissions []db.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

// Add more CRUD methods as needed...
