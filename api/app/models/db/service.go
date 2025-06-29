package db

import (
	models_core "base_lara_go_project/app/core/models"

	"gorm.io/gorm"
)

type Service struct {
	models_core.BaseModel
	gorm.Model
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:varchar(500)" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	CategoryID  uint      `gorm:"index" json:"category_id"`
	Category    *Category `gorm:"foreignKey:CategoryID" json:"category"`
}

func (Service) TableName() string {
	return "services"
}

func (service *Service) AfterFind(tx *gorm.DB) (err error) {
	service.BaseModelData = *models_core.NewBaseModel()
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

func NewService() *Service {
	return &Service{
		BaseModel: models_core.BaseModel{
			BaseModelData: *models_core.NewBaseModel(),
		},
	}
}
