package models

import (
	"base_lara_go_project/app/core"

	"gorm.io/gorm"
)

type Category struct {
	core.BaseModel
	gorm.Model
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	Services    []Service `gorm:"foreignkey:CategoryID" json:"services"`
}

// TableName returns the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

func (category *Category) AfterFind(tx *gorm.DB) (err error) {
	category.BaseModel = *core.NewBaseModel()
	category.BaseModel.Set("id", category.ID)
	category.BaseModel.Set("name", category.Name)
	category.BaseModel.Set("description", category.Description)
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

// GetServicesCount returns the number of services in this category
func (category *Category) GetServicesCount() int {
	return len(category.Services)
}
