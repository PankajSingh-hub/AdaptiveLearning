package database

import (
	"log"

	"adaptive-assessment/internal/config"
	"adaptive-assessment/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if cfg.AutoMigrate {
		err = db.AutoMigrate(
			&models.User{}, &models.Topic{}, &models.Question{}, &models.QuestionOption{},
			&models.Assessment{}, &models.InterviewSession{}, &models.SessionQuestion{}, &models.Response{},
		)
		if err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	}
	return db
}