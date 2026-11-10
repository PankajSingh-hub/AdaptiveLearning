package database

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

	"adaptive-assessment/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed fills an empty database with demo data. It is safe to run on every start.
func Seed(db *gorm.DB) {
	seedAdmin(db)
	// The demo candidate is only created when you opt in (local development).
	if os.Getenv("SEED_DEMO_CANDIDATE") == "true" {
		seedUser(db, "Sample Candidate", "candidate@example.com", env("CANDIDATE_PASSWORD", "Candidate@123"), models.RoleCandidate)
	}
	seedQuestions(db)
	seedAssessments(db)
	seedCodingAssessment(db)
}

// seedAdmin never uses a well-known password: set ADMIN_PASSWORD, or a random one is generated and logged once.
func seedAdmin(db *gorm.DB) {
	const email = "admin@example.com"
	var count int64
	db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return
	}
	pw := os.Getenv("ADMIN_PASSWORD")
	if pw == "" {
		pw = randomPassword()
		log.Printf("ADMIN ACCOUNT CREATED: email=%s password=%s (shown once, store it safely)", email, pw)
	}
	seedUser(db, "Admin", email, pw, models.RoleAdmin)
}

func randomPassword() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("seed: random password failed: %v", err)
	}
	return hex.EncodeToString(b)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func seedUser(db *gorm.DB, name, email, password string, role models.Role) {
	var count int64
	db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("seed: hashing failed: %v", err)
	}
	if err := db.Create(&models.User{Name: name, Email: email, PasswordHash: string(hash), Role: role}).Error; err != nil {
		log.Fatalf("seed: creating %s failed: %v", email, err)
	}
	log.Printf("seeded %s user: %s", role, email)
}

func seedQuestions(db *gorm.DB) {
	var count int64
	db.Model(&models.Question{}).Count(&count)
	if count > 0 {
		return
	}
	for _, s := range seedQuestionData {
		var topic models.Topic
		if err := db.Where(models.Topic{Name: s.topic}).FirstOrCreate(&topic).Error; err != nil {
			log.Fatalf("seed: topic failed: %v", err)
		}
		opts := make([]models.QuestionOption, 0, len(s.opts))
		for i, text := range s.opts {
			opts = append(opts, models.QuestionOption{OptionText: text, IsCorrect: i == s.correct})
		}
		q := models.Question{
			QuestionText: s.text, TopicID: topic.ID, Category: s.category, Difficulty: s.difficulty,
			QuestionType: "MCQ", Explanation: s.explanation, TimeLimitSec: 60, Active: true, Options: opts,
		}
		if err := db.Create(&q).Error; err != nil {
			log.Fatalf("seed: question failed: %v", err)
		}
	}
	log.Printf("seeded %d questions", len(seedQuestionData))
}

func seedAssessments(db *gorm.DB) {
	var count int64
	db.Model(&models.Assessment{}).Count(&count)
	if count > 0 {
		return
	}
	list := []models.Assessment{
		{Title: "Full-Stack Fundamentals", Description: "Adaptive mix of DSA, Java, DBMS, OS and networking.",
			TotalQuestions: 10, DurationMinutes: 30, ShowExplanations: false, Active: true},
		{Title: "Quick Screening", Description: "A short 5-question adaptive screening.",
			TotalQuestions: 5, DurationMinutes: 10, ShowExplanations: true, Active: true},
	}
	if err := db.Create(&list).Error; err != nil {
		log.Fatalf("seed: assessments failed: %v", err)
	}
	log.Println("seeded 2 assessments")
}
func seedCodingAssessment(db *gorm.DB) {
	var count int64
	db.Model(&models.Assessment{}).Where("assessment_type = ?", "CODING").Count(&count)

	if count > 0 {
		return
	}

	a := models.Assessment{
		Title:           "DSA Coding Challenge",
		AssessmentType:  "CODING",
		Active:          true,
		Description:     "10 randomly chosen classic DSA problems. Solve in C++, Java or Python with progressive hints.",
		TotalQuestions:  10,
		DurationMinutes: 90,
	}

	if err := db.Create(&a).Error; err != nil {
		log.Fatalf("seed: coding assessment failed: %v", err)
	}

	log.Println("seeded coding assessment")
}