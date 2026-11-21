package service

import (
	"log"
	"strings"
	"time"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	dayFmt              = "2006-01-02"
	streakLookbackDays  = 120
)

type ProfileService struct{ db *gorm.DB }

func NewProfileService(db *gorm.DB) *ProfileService { return &ProfileService{db: db} }

func (s *ProfileService) Get(userID uint, tz string) (*dto.ProfileDTO, error) {
	var u models.User
	if err := s.db.First(&u, userID).Error; err != nil {
		return nil, NotFound("user not found")
	}
	out := &dto.ProfileDTO{
		ID: u.ID, Name: u.Name, Email: u.Email, Role: string(u.Role), JoinedAt: u.CreatedAt,
		EmailVerified: u.EmailVerifiedAt != nil, IsDemo: u.Email == DemoEmail,
	}
	if err := s.db.Model(&models.InterviewReport{}).Where("candidate_id = ?", userID).
		Count(&out.InterviewsCompleted).Error; err != nil {
		return nil, err
	}
	var answered struct{ N int64 }
	if err := s.db.Model(&models.InterviewSession{}).
		Select("COALESCE(CAST(SUM(answered_count) AS SIGNED),0) AS n").
		Where("candidate_id = ?", userID).Scan(&answered).Error; err != nil {
		return nil, err
	}
	out.QuestionsAnswered = answered.N
	if err := s.db.Table("coding_session_problems p").
		Joins("JOIN coding_sessions s ON s.id = p.session_id").
		Where("s.candidate_id = ? AND p.status = ?", userID, "SOLVED").
		Count(&out.CodingSolved).Error; err != nil {
		return nil, err
	}
	if st, err := s.Streak(userID, tz); err == nil {
		out.CurrentStreak = st.Current
	}
	return out, nil
}

func (s *ProfileService) UpdateName(userID uint, name string) (*dto.UserDTO, error) {
	var u models.User
	if err := s.db.First(&u, userID).Error; err != nil {
		return nil, NotFound("user not found")
	}
	if u.Email == DemoEmail {
		return nil, Forbidden("the shared demo account cannot be changed")
	}
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, BadRequest("name must be at least 2 characters")
	}
	if err := s.db.Model(&u).Update("name", name).Error; err != nil {
		return nil, err
	}
	return toUserDTO(u), nil
}

// ChangePassword failures use 400, never 401: the frontend treats 401 as "session expired" and logs out.
func (s *ProfileService) ChangePassword(userID uint, req dto.ChangePasswordRequest) error {
	var u models.User
	if err := s.db.First(&u, userID).Error; err != nil {
		return NotFound("user not found")
	}
	if u.Email == DemoEmail {
		return Forbidden("the shared demo account's password cannot be changed")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.CurrentPassword)) != nil {
		return BadRequest("current password is incorrect")
	}
	if req.NewPassword != req.ConfirmPassword {
		return BadRequest("new passwords do not match")
	}
	if req.NewPassword == req.CurrentPassword {
		return BadRequest("the new password must be different from the current one")
	}
	if err := validatePassword(req.NewPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.db.Model(&u).Update("password_hash", string(hash)).Error; err != nil {
		return err
	}
	log.Printf("password changed: user=%d", userID)
	return nil
}

// Streak reads quiz answers and code submissions and counts the days with activity in the user's timezone.
func (s *ProfileService) Streak(userID uint, tz string) (*dto.StreakDTO, error) {
	loc := loadLocation(tz)
	now := time.Now().In(loc)
	since := now.AddDate(0, 0, -streakLookbackDays)
	var rows []struct{ TakenAt time.Time }
	err := s.db.Raw(`
		SELECT r.created_at AS taken_at FROM responses r
		  JOIN session_questions sq ON sq.id = r.session_question_id
		  JOIN interview_sessions s ON s.id = sq.session_id
		 WHERE s.candidate_id = ? AND r.created_at >= ?
		UNION ALL
		SELECT cs.created_at AS taken_at FROM coding_submissions cs
		  JOIN coding_session_problems p ON p.id = cs.session_problem_id
		  JOIN coding_sessions s ON s.id = p.session_id
		 WHERE s.candidate_id = ? AND cs.created_at >= ?`,
		userID, since, userID, since).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	active := map[string]bool{}
	for _, r := range rows {
		active[r.TakenAt.In(loc).Format(dayFmt)] = true
	}
	return buildStreak(active, now), nil
}

// buildStreak is pure (no DB), so it is unit tested. Today without activity yet does not break the streak.
func buildStreak(active map[string]bool, now time.Time) *dto.StreakDTO {
	at := func(back int) string { return now.AddDate(0, 0, -back).Format(dayFmt) }
	start := 0
	if !active[at(0)] {
		start = 1
	}
	current := 0
	for i := start; i < streakLookbackDays && active[at(i)]; i++ {
		current++
	}
	days := make([]dto.StreakDay, 0, 7)
	for i := 6; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		days = append(days, dto.StreakDay{Date: d.Format(dayFmt), Label: d.Format("Mon"), Active: active[d.Format(dayFmt)], Today: i == 0})
	}
	return &dto.StreakDTO{Current: current, Days: days}
}

func loadLocation(tz string) *time.Location {
	if tz == "" || len(tz) > 64 {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}