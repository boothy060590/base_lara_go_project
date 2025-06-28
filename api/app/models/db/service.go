package db

import (
	"base_lara_go_project/app/core"

	"gorm.io/gorm"
)

type Service struct {
	core.BaseModel
	gorm.Model
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	CategoryID  uint      `json:"category_id"`
	Category    *Category `gorm:"foreignKey:CategoryID" json:"category"`
}

func (Service) TableName() string {
	return "services"
}

func (service *Service) AfterFind(tx *gorm.DB) (err error) {
	service.BaseModelData = *core.NewBaseModel()
	service.BaseModelData.Set("id", service.ID)
	service.BaseModelData.Set("name", service.Name)
	service.BaseModelData.Set("description", service.Description)
	service.BaseModelData.Set("category_id", service.CategoryID)
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
