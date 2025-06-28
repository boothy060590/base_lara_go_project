package repositories

import (
	"base_lara_go_project/app/models/db"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindByID(id uint) (*db.Category, error) {
	var category db.Category
	err := r.db.First(&category, id).Error
	return &category, err
}

func (r *CategoryRepository) All() ([]db.Category, error) {
	var categories []db.Category
	err := r.db.Find(&categories).Error
	return categories, err
}

// Add more CRUD methods as needed...
