package database

import (
	"fmt"

	"github.com/endyapina/muzzapp/internal/config"
	"github.com/endyapina/muzzapp/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Init initializes the database connection using the provided configuration.
//
// This function uses GORM (an ORM library for Go) to quickly set up
// a connection to a MySQL database. GORM makes it easy to work with
// models and perform CRUD operations without writing raw SQL.
//
// While GORM is very convenient for prototyping and small- to medium-sized
// apps, in high-scale production systems it is advised to use raw SQL queries
// or a lightweight database library. This can give you finer control over
// performance, query optimization, and transaction handling.
func Init(config *config.AppConfig) (*gorm.DB, error) {
	if config == nil {
		return nil, fmt.Errorf("missing database config")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// automatically create or update the schema for the given models.
	// in production you may want to manage schema migrations
	// explicitly using a tool like golang-migrate or Flyway
	db.AutoMigrate(&models.Decision{})
	return db, nil
}
