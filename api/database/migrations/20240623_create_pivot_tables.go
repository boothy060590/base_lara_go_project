package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var CreatePivotTables = &gormigrate.Migration{
	ID: "20240623_create_pivot_tables",
	Migrate: func(tx *gorm.DB) error {
		// Create user_roles and role_permissions as join tables
		type UserRole struct {
			UserID uint `gorm:"primaryKey"`
			RoleID uint `gorm:"primaryKey"`
		}
		type RolePermission struct {
			RoleID       uint `gorm:"primaryKey"`
			PermissionID uint `gorm:"primaryKey"`
		}
		if err := tx.AutoMigrate(&UserRole{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&RolePermission{}); err != nil {
			return err
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		if err := tx.Migrator().DropTable("user_roles"); err != nil {
			return err
		}
		if err := tx.Migrator().DropTable("role_permissions"); err != nil {
			return err
		}
		return nil
	},
}
