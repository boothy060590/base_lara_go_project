package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var CreatePermissions = &gormigrate.Migration{
	ID: "20240623_create_permissions",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
			CREATE TABLE IF NOT EXISTS permissions (
				id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
				created_at DATETIME NULL,
				updated_at DATETIME NULL,
				deleted_at DATETIME NULL,
				name VARCHAR(64) NOT NULL UNIQUE,
				description VARCHAR(255) NULL,
				INDEX idx_permissions_deleted_at (deleted_at)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
		`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("permissions")
	},
}
