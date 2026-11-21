package dto

import "time"

type ProfileDTO struct {
	ID                  uint      `json:"id"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	Role                string    `json:"role"`
	JoinedAt            time.Time `json:"joinedAt"`
	EmailVerified       bool      `json:"emailVerified"`
	IsDemo              bool      `json:"isDemo"`
	InterviewsCompleted int64     `json:"interviewsCompleted"`
	QuestionsAnswered   int64     `json:"questionsAnswered"`
	CodingSolved        int64     `json:"codingSolved"`
	CurrentStreak       int       `json:"currentStreak"`
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type StreakDay struct {
	Date   string `json:"date"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
	Today  bool   `json:"today"`
}

type StreakDTO struct {
	Current int         `json:"current"`
	Days    []StreakDay `json:"days"` // the last 7 days, oldest first
}