package models

import (
	"time"
)

// Service represents a service in the system
type Service struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Slug        string     `json:"slug"`
	Price       float64    `json:"price"`
	CategoryID  uint       `json:"category_id"`
	Category    Category   `json:"category"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
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
