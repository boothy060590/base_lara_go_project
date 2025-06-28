package repositories

import (
	"base_lara_go_project/app/models/db"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByID(id uint) (*db.Role, error) {
	var role db.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

func (r *RoleRepository) FindByName(name string) (*db.Role, error) {
	var role db.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *RoleRepository) All() ([]db.Role, error) {
	var roles []db.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// Add more CRUD methods as needed...
