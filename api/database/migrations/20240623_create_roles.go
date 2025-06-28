package migrations

import (
	db "base_lara_go_project/app/models/db"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var CreateRoles = &gormigrate.Migration{
	ID: "20240623_create_roles",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&db.Role{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("roles")
	},
}
