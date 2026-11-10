package models

import "time"

type CodingSession struct {
	ID           uint `gorm:"primaryKey"`
	CandidateID  uint
	AssessmentID uint
	Status       SessionStatus
	Score        int
	StartedAt    time.Time
	ExpiresAt    time.Time
	CompletedAt  *time.Time
		Warnings    int
		Terminated  bool `gorm:"column:is_terminated"`
}

// ProblemSlug refers to a problem in the in-code catalog (internal/coding).
type CodingSessionProblem struct {
	ID          uint `gorm:"primaryKey"`
	SessionID   uint
	ProblemSlug string
	OrderNo     int
	Status      string // OPEN, SOLVED, FAILED, SKIPPED
	Attempts    int
	Failures    int
	Language    string
	Points      int
	RuntimeMs   int
	SolvedAt    *time.Time
}

type CodingSubmission struct {
	ID               uint `gorm:"primaryKey"`
	SessionProblemID uint
	Language         string
	Code             string
	Verdict          string
	Passed           int
	Total            int
	RuntimeMs        int
	CreatedAt        time.Time
}