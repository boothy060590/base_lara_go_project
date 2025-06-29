package db

import (
	models_core "base_lara_go_project/app/core/models"

	"gorm.io/gorm"
)

type Category struct {
	models_core.BaseModelData
	gorm.Model
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	Description string     `gorm:"type:varchar(500)" json:"description"`
	ParentID    *uint      `gorm:"index" json:"parent_id"`
	Parent      *Category  `gorm:"foreignKey:ParentID" json:"parent"`
	Children    []Category `gorm:"foreignKey:ParentID" json:"children"`
	Services    []*Service `gorm:"foreignkey:CategoryID" json:"services"`
}

func (Category) TableName() string {
	return "categories"
}

func (category *Category) AfterFind(tx *gorm.DB) (err error) {
	category.BaseModelData = *models_core.NewBaseModel()
	category.BaseModelData.Set("id", category.ID)
	category.BaseModelData.Set("name", category.Name)
	category.BaseModelData.Set("description", category.Description)
	return nil
}

func (category *Category) AfterCreate(tx *gorm.DB) (err error) {
	category.AfterFind(tx)
	return nil
}

func (category *Category) AfterUpdate(tx *gorm.DB) (err error) {
	category.AfterFind(tx)
	return nil
}

func (category *Category) GetServicesCount() int {
	return len(category.Services)
}

// NewCategory creates a new category with initialized base model data
func NewCategory() *Category {
	return &Category{
		BaseModelData: *models_core.NewBaseModel(),
	}
}
