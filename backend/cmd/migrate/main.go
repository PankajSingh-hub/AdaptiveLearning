package main

import (
	"log"

	"adaptive-assessment/internal/config"
	"adaptive-assessment/internal/database"
)

func main() {
	cfg := config.Load()
	if err := database.Migrate(cfg.DSN); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations are up to date")
}