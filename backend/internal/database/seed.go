package database

import (
	"log"
	"os"

	"adaptive-assessment/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed fills an empty database with demo data. It is safe to run on every start.
func Seed(db *gorm.DB) {
	seedUser(db, "Admin", "admin@example.com", env("ADMIN_PASSWORD", "Admin@12345"), models.RoleAdmin)
	seedUser(db, "Sample Candidate", "candidate@example.com", env("CANDIDATE_PASSWORD", "Candidate@123"), models.RoleCandidate)
	seedQuestions(db)
	seedAssessments(db)
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