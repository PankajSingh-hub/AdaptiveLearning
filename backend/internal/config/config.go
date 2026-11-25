package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DSN            string
	JWTSecret      string
	JWTExpiryHrs   int
	AllowedOrigin  string
	RunMigrations  bool
	DemoLogin      bool
	Production     bool
	CodingExecutor string // docker | local | off
}

func Load() *Config {
	_ = godotenv.Load() // .env is optional (e.g., in Docker)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	prod := strings.EqualFold(get("APP_ENV", "development"), "production")
	if prod && len(secret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters when APP_ENV=production")
	}
	mode := strings.ToLower(get("CODING_EXECUTOR", "off"))
	if mode == "local" && prod {
		log.Fatal("CODING_EXECUTOR=local runs code without a sandbox and is not allowed in production")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		get("DB_USER", "root"), os.Getenv("DB_PASSWORD"),
		get("DB_HOST", "localhost"), get("DB_PORT", "3306"), get("DB_NAME", "adaptive_assessment"))
	if tls := os.Getenv("DB_TLS"); tls != "" { // true | skip-verify | preferred, for managed MySQL
		dsn += "&tls=" + tls
	}
	hrs, err := strconv.Atoi(get("JWT_EXPIRY_HOURS", "8"))
	if err != nil || hrs < 1 {
		hrs = 8
	}
	return &Config{
		Port:           get("PORT", "8080"),
		DSN:            dsn,
		JWTSecret:      secret,
		JWTExpiryHrs:   hrs,
		AllowedOrigin:  get("ALLOWED_ORIGIN", "http://localhost:5173"),
		RunMigrations:  get("RUN_MIGRATIONS", "true") == "true",
		DemoLogin:      get("DEMO_LOGIN", "false") == "true",
		Production:     prod,
		CodingExecutor: mode,
	}
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}