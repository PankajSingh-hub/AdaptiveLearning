package models

import "time"

type Role string
type Difficulty string
type SessionStatus string

const (
	RoleCandidate Role = "CANDIDATE"
	RoleAdmin     Role = "ADMIN"

	Easy   Difficulty = "EASY"
	Medium Difficulty = "MEDIUM"
	Hard   Difficulty = "HARD"

	InProgress SessionStatus = "IN_PROGRESS"
	Completed  SessionStatus = "COMPLETED"
	Expired    SessionStatus = "EXPIRED"
)

type User struct {
	ID           uint `gorm:"primaryKey"`
	Name         string
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Topic struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type Question struct {
	ID           uint `gorm:"primaryKey"`
	QuestionText string
	TopicID      uint
	Topic        Topic
	Category     string
	Difficulty   Difficulty
	QuestionType string
	Explanation  string
	TimeLimitSec int
	Active       bool
	Options      []QuestionOption
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type QuestionOption struct {
	ID         uint `gorm:"primaryKey"`
	QuestionID uint
	OptionText string
	IsCorrect  bool
}

type Assessment struct {
	ID               uint `gorm:"primaryKey"`
	Title            string
	Description      string
	TotalQuestions   int
	DurationMinutes  int
	ShowExplanations bool
	Active           bool
	AssessmentType   string `gorm:"default:MCQ"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type InterviewSession struct {
	ID                 uint `gorm:"primaryKey"`
	CandidateID        uint
	AssessmentID       uint
	Status             SessionStatus
	CurrentDifficulty  Difficulty
	ConsecutiveCorrect int
	ConsecutiveWrong   int
	AnsweredCount      int
	CorrectCount       int
	TotalScore         int
	StartedAt          time.Time
	ExpiresAt          time.Time
	CompletedAt        *time.Time
		Warnings           int
	Terminated         bool
}

type SessionQuestion struct {
	ID         uint `gorm:"primaryKey"`
	SessionID  uint
	QuestionID uint
	OrderNo    int
	Difficulty Difficulty
	ServedAt   time.Time
	AnsweredAt *time.Time
}

type Response struct {
	ID                uint `gorm:"primaryKey"`
	SessionQuestionID uint
	SelectedOptionID  uint
	IsCorrect         bool
	PointsAwarded     int
	TimeTakenSec      int
	CreatedAt         time.Time
}