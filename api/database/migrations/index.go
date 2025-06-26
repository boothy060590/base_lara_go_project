package migrations

import "github.com/go-gormigrate/gormigrate/v2"

func AllMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		CreateUsers,
		CreateRoles,
		CreatePermissions,
		CreatePivotTables,
	}
}
