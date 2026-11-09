package dto

import "time"

type OptionInput struct {
	Text    string `json:"text" binding:"required,max=500"`
	Correct bool   `json:"correct"`
}

type QuestionRequest struct {
	Text         string        `json:"text" binding:"required,min=10"`
	Topic        string        `json:"topic" binding:"required,max=100"`
	Category     string        `json:"category" binding:"required"`
	Difficulty   string        `json:"difficulty" binding:"required,oneof=EASY MEDIUM HARD"`
	Explanation  string        `json:"explanation"`
	TimeLimitSec int           `json:"timeLimitSec" binding:"omitempty,min=10,max=600"`
	Active       *bool         `json:"active"`
	Options      []OptionInput `json:"options" binding:"required,min=2,max=6,dive"`
}

// Admin view only: includes the correct flag. Never use this for candidates.
type AdminOptionDTO struct {
	ID      uint   `json:"id"`
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
}

type AdminQuestionDTO struct {
	ID           uint             `json:"id"`
	Text         string           `json:"text"`
	Topic        string           `json:"topic"`
	Category     string           `json:"category"`
	Difficulty   string           `json:"difficulty"`
	QuestionType string           `json:"questionType"`
	Explanation  string           `json:"explanation"`
	TimeLimitSec int              `json:"timeLimitSec"`
	Active       bool             `json:"active"`
	Options      []AdminOptionDTO `json:"options"`
	CreatedAt    time.Time        `json:"createdAt"`
}

type PagedQuestions struct {
	Items      []AdminQuestionDTO `json:"items"`
	Page       int                `json:"page"`
	Size       int                `json:"size"`
	TotalItems int64              `json:"totalItems"`
	TotalPages int                `json:"totalPages"`
}

type AssessmentRequest struct {
	Title            string `json:"title" binding:"required,min=3,max=150"`
	Description      string `json:"description" binding:"max=500"`
	TotalQuestions   int    `json:"totalQuestions" binding:"required,min=1,max=50"`
	DurationMinutes  int    `json:"durationMinutes" binding:"required,min=1,max=240"`
	ShowExplanations bool   `json:"showExplanations"`
	Active           *bool  `json:"active"`
}

type AssessmentDTO struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	TotalQuestions   int       `json:"totalQuestions"`
	DurationMinutes  int       `json:"durationMinutes"`
	ShowExplanations bool      `json:"showExplanations"`
	Active           bool      `json:"active"`
		Type             string    `json:"type"`
	CreatedAt        time.Time `json:"createdAt"`
}