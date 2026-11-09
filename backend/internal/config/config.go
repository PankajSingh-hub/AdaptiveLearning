package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DSN           string
	JWTSecret     string
	JWTExpiryHrs  int
	AllowedOrigin string
	AutoMigrate   bool
}

func Load() *Config {
	_ = godotenv.Load() // .env is optional (e.g., in Docker)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		get("DB_USER", "root"), os.Getenv("DB_PASSWORD"),
		get("DB_HOST", "localhost"), get("DB_PORT", "3306"), get("DB_NAME", "adaptive_assessment"))
	hrs, _ := strconv.Atoi(get("JWT_EXPIRY_HOURS", "24"))
	return &Config{
		Port:          get("PORT", "8080"),
		DSN:           dsn,
		JWTSecret:     secret,
		JWTExpiryHrs:  hrs,
		AllowedOrigin: get("ALLOWED_ORIGIN", "http://localhost:5173"),
		AutoMigrate:   get("AUTO_MIGRATE", "true") == "true",
	}
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}