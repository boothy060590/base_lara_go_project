package providers

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
	database_core "base_lara_go_project/app/core/database"
	"base_lara_go_project/config"
	"base_lara_go_project/database/migrations"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var DB *gorm.DB

type DatabaseServiceProvider struct{}

func NewDatabaseServiceProvider() *DatabaseServiceProvider {
	return &DatabaseServiceProvider{}
}

func (p *DatabaseServiceProvider) Register() error {
	factory := database_core.NewDatabaseProviderFactory(app_core.App)
	if err := factory.RegisterFromConfig(config.DatabaseConfig()); err != nil {
		return fmt.Errorf("failed to register database provider: %w", err)
	}
	return nil
}

func RegisterDatabase() error {
	provider := NewDatabaseServiceProvider()
	return provider.Register()
}

func RunMigrations() {
	m := gormigrate.New(DB, gormigrate.DefaultOptions, migrations.AllMigrations())
	if err := m.Migrate(); err != nil {
		panic("Could not migrate: " + err.Error())
	}
}
