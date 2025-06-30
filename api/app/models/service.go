package models

import (
	"time"

	"gorm.io/gorm"
)

// Service represents a service in the system
type Service struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Slug        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	CategoryID  uint           `json:"category_id"`
	Category    Category       `json:"category" gorm:"foreignKey:CategoryID"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for the Service model
func (Service) TableName() string {
	return "services"
}

// NewService creates a new service instance
func NewService() *Service {
	return &Service{
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetName returns the service name
func (s *Service) GetName() string {
	return s.Name
}

// GetDescription returns the service description
func (s *Service) GetDescription() string {
	return s.Description
}

// GetSlug returns the service slug
func (s *Service) GetSlug() string {
	return s.Slug
}

// GetPrice returns the service price
func (s *Service) GetPrice() float64 {
	return s.Price
}

// IsServiceActive checks if the service is active
func (s *Service) IsServiceActive() bool {
	return s.IsActive
}

// GetCategory returns the service category
func (s *Service) GetCategory() Category {
	return s.Category
}
