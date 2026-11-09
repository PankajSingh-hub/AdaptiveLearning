package models

import "time"

// InterviewReport is a stored snapshot generated once when a session ends.
// List-like fields are stored as JSON text.
type InterviewReport struct {
	ID                    uint `gorm:"primaryKey"`
	SessionID             uint
	CandidateID           uint
	AssessmentID          uint
	TotalQuestions        int
	Attempted             int
	CorrectCount          int
	Accuracy              float64
	Score                 int
	MaxScore              int
	AvgResponseSec        float64
	PerformanceLevel      string
	StrongTopics          string
	WeakTopics            string
	TopicBreakdown        string
	DifficultyBreakdown   string
	DifficultyProgression string
	Recommendations       string
	CreatedAt             time.Time
}