package models

import (
	"time"

	"gorm.io/gorm"
)

// Category represents a category in the system
type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Slug        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	ParentID    *uint          `json:"parent_id"`
	Parent      *Category      `json:"parent" gorm:"foreignKey:ParentID"`
	Children    []Category     `json:"children" gorm:"foreignKey:ParentID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

// NewCategory creates a new category instance
func NewCategory() *Category {
	return &Category{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetName returns the category name
func (c *Category) GetName() string {
	return c.Name
}

// GetDescription returns the category description
func (c *Category) GetDescription() string {
	return c.Description
}

// GetSlug returns the category slug
func (c *Category) GetSlug() string {
	return c.Slug
}

// HasParent checks if the category has a parent
func (c *Category) HasParent() bool {
	return c.ParentID != nil
}

// HasChildren checks if the category has children
func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}
