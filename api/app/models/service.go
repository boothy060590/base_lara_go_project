package models

import (
	"base_lara_go_project/app/core"

	"gorm.io/gorm"
)

type Service struct {
	core.BaseModel
	gorm.Model
	Name        string   `gorm:"type:varchar(255);not null" json:"name"`
	Description string   `gorm:"type:varchar(255)" json:"description"`
	CategoryID  uint     `json:"category_id"`
	Category    Category `gorm:"foreignKey:CategoryID" json:"category"`
}

// TableName returns the table name for the Service model
func (Service) TableName() string {
	return "services"
}

func (service *Service) AfterFind(tx *gorm.DB) (err error) {
	service.BaseModel = *core.NewBaseModel()
	service.BaseModel.Set("id", service.ID)
	service.BaseModel.Set("name", service.Name)
	service.BaseModel.Set("description", service.Description)
	service.BaseModel.Set("category_id", service.CategoryID)
	return nil
}

func (service *Service) AfterCreate(tx *gorm.DB) (err error) {
	service.AfterFind(tx)
	return nil
}

func (service *Service) AfterUpdate(tx *gorm.DB) (err error) {
	service.AfterFind(tx)
	return nil
}
