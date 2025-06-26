package migrations

import (
	"base_lara_go_project/app/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var CreateRoles = &gormigrate.Migration{
	ID: "20240623_create_roles",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&models.Role{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("roles")
	},
}
