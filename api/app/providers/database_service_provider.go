package providers

import (
	"fmt"
	"log"
	"os"

	"base_lara_go_project/app/core"
	"base_lara_go_project/database/migrations"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// DatabaseProvider implements the core DatabaseInterface
type DatabaseProvider struct {
	db *gorm.DB
}

// NewDatabaseProvider creates a new database provider
func NewDatabaseProvider(db *gorm.DB) *DatabaseProvider {
	return &DatabaseProvider{db: db}
}

// Basic operations that are used by the facade
func (d *DatabaseProvider) Create(value interface{}) error {
	return d.db.Create(value).Error
}

func (d *DatabaseProvider) First(dest interface{}, conds ...interface{}) error {
	return d.db.First(dest, conds...).Error
}

func (d *DatabaseProvider) Find(dest interface{}, conds ...interface{}) error {
	return d.db.Find(dest, conds...).Error
}

func (d *DatabaseProvider) Save(value interface{}) error {
	return d.db.Save(value).Error
}

func (d *DatabaseProvider) Delete(value interface{}, conds ...interface{}) error {
	return d.db.Delete(value, conds...).Error
}

// Query builder methods that are used by the facade
func (d *DatabaseProvider) Table(tableName string) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Table(tableName)}
}

func (d *DatabaseProvider) Where(query interface{}, args ...interface{}) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Where(query, args...)}
}

func (d *DatabaseProvider) Preload(query string, args ...interface{}) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Preload(query, args...)}
}

func (d *DatabaseProvider) Model(value interface{}) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Model(value)}
}

// Additional methods that might be needed by the facade
func (d *DatabaseProvider) Order(value interface{}) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Order(value)}
}

func (d *DatabaseProvider) Limit(limit int) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Limit(limit)}
}

func (d *DatabaseProvider) Offset(offset int) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Offset(offset)}
}

// Additional methods required by the interface
func (d *DatabaseProvider) Or(query interface{}, args ...interface{}) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Or(query, args...)}
}

func (d *DatabaseProvider) Joins(query string, args ...interface{}) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Joins(query, args...)}
}

func (d *DatabaseProvider) Transaction(fc func(tx core.DatabaseInterface) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		txProvider := &DatabaseProvider{db: tx}
		return fc(txProvider)
	})
}

func (d *DatabaseProvider) Raw(sql string, values ...interface{}) core.DatabaseInterface {
	return &DatabaseProvider{db: d.db.Raw(sql, values...)}
}

func (d *DatabaseProvider) Exec(sql string, values ...interface{}) error {
	return d.db.Exec(sql, values...).Error
}

func (d *DatabaseProvider) Migrate() error {
	// This would be implemented to run migrations
	// For now, we'll return nil as migrations are handled separately
	return nil
}

func RegisterDatabase() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	DbHost := os.Getenv("DB_HOST")
	DbUser := os.Getenv("DB_USER")
	DbPassword := os.Getenv("DB_PASSWORD")
	DbName := os.Getenv("DB_NAME")
	DbPort := os.Getenv("DB_PORT")

	DSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)

	DB, err = gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		fmt.Println("Cannot connect to database using GORM v2")
		log.Fatal("connection error:", err)
	} else {
		fmt.Println("We are connected to the database using GORM v2")
	}

	// Set up the global database instance with our provider
	core.DatabaseInstance = NewDatabaseProvider(DB)
}

func RunMigrations() {
	m := gormigrate.New(DB, gormigrate.DefaultOptions, migrations.AllMigrations())
	if err := m.Migrate(); err != nil {
		panic("Could not migrate: " + err.Error())
	}
}
