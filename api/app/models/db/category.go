package db

import (
	"base_lara_go_project/app/core"

	"gorm.io/gorm"
)

type Category struct {
	core.BaseModelData
	gorm.Model
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	Description string     `gorm:"type:varchar(255)" json:"description"`
	Services    []*Service `gorm:"foreignkey:CategoryID" json:"services"`
}

func (Category) TableName() string {
	return "categories"
}

func (category *Category) AfterFind(tx *gorm.DB) (err error) {
	category.BaseModelData = *core.NewBaseModel()
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
