package service

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
)

// Not a valid bcrypt hash, so nobody can ever sign in to a demo account with a password.
// Only the demo endpoint can issue a session for one.
const demoNoPassword = "!demo-account-has-no-password"

var demoPurge struct {
	sync.Mutex
	last time.Time
}

// DemoLogin creates a brand-new temporary candidate, so every demo visit starts empty.
func (s *AuthService) DemoLogin() (*dto.LoginResponse, error) {
	s.purgeOldDemoUsers()

	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	u := models.User{
		Name:         "Demo Candidate",
		Email:        "demo-" + hex.EncodeToString(b) + "@demo.invalid", // .invalid can never be a real address
		PasswordHash: demoNoPassword,
		Role:         models.RoleCandidate,
		IsDemo:       true,
	}
	if err := s.db.Create(&u).Error; err != nil {
		return nil, err
	}
	token, err := s.generateToken(u)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{Token: token, User: *toUserDTO(u)}, nil
}

// purgeOldDemoUsers deletes demo accounts older than 24 hours, with all their sessions and results.
// It runs at most once every 10 minutes.
func (s *AuthService) purgeOldDemoUsers() {
	demoPurge.Lock()
	defer demoPurge.Unlock()
	if time.Since(demoPurge.last) < 10*time.Minute {
		return
	}
	demoPurge.last = time.Now()

	const old = "SELECT id FROM users WHERE is_demo = 1 AND created_at < NOW() - INTERVAL 24 HOUR"
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Deleting sessions cascades to their questions, responses, reports and coding rows.
		if err := tx.Exec("DELETE FROM interview_sessions WHERE candidate_id IN (" + old + ")").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM coding_sessions WHERE candidate_id IN (" + old + ")").Error; err != nil {
			return err
		}
		res := tx.Exec("DELETE FROM users WHERE is_demo = 1 AND created_at < NOW() - INTERVAL 24 HOUR")
		if res.Error == nil && res.RowsAffected > 0 {
			log.Printf("demo cleanup: removed %d old demo account(s)", res.RowsAffected)
		}
		return res.Error
	})
	if err != nil {
		log.Printf("demo cleanup failed: %v", err)
	}
}