package providers

import (
	"fmt"
	"log"

	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/db"
	"base_lara_go_project/config"
	"base_lara_go_project/database/migrations"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func RegisterDatabase() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbConfig := config.DatabaseConfig()
	defaultConn := dbConfig["default"].(string)
	connections := dbConfig["connections"].(map[string]interface{})
	connectionConfig := connections[defaultConn].(map[string]interface{})

	DbHost := connectionConfig["host"].(string)
	DbUser := connectionConfig["username"].(string)
	DbPassword := connectionConfig["password"].(string)
	DbName := connectionConfig["database"].(string)
	DbPort := connectionConfig["port"].(string)

	DSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)

	DB, err = gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		fmt.Println("Cannot connect to database using GORM v2")
		log.Fatal("connection error:", err)
	} else {
		fmt.Println("We are connected to the database using GORM v2")
	}

	// Set up the global database instance with our provider
	core.DatabaseInstance = core.NewDatabaseProvider(DB)

	// Register cacheable models for automatic cache invalidation
	core.RegisterCacheableModel(DB, &db.User{})
}

func RunMigrations() {
	m := gormigrate.New(DB, gormigrate.DefaultOptions, migrations.AllMigrations())
	if err := m.Migrate(); err != nil {
		panic("Could not migrate: " + err.Error())
	}
}
