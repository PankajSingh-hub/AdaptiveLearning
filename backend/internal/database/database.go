package database

import (
	"log"

	"adaptive-assessment/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connect runs the SQL migrations (unless disabled) and opens the GORM connection.
// GORM's AutoMigrate is deliberately NOT used: it cannot create the ENUMs and unique
// constraints this schema relies on.
func Connect(cfg *config.Config) *gorm.DB {
	if cfg.RunMigrations {
		if err := Migrate(cfg.DSN); err != nil {
			log.Fatalf("database migration failed: %v", err)
		}
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	return db
}