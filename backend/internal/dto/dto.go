package dto

import "time"

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserDTO struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// ---- Interview ----

type StartRequest struct {
	AssessmentID uint `json:"assessmentId" binding:"required"`
}

type OptionDTO struct {
	ID   uint   `json:"id"`
	Text string `json:"text"`
} // deliberately NO isCorrect field

type QuestionDTO struct {
	SessionQuestionID uint        `json:"sessionQuestionId"`
	Number            int         `json:"number"`
	Total             int         `json:"total"`
	Text              string      `json:"text"`
	Topic             string      `json:"topic"`
	Difficulty        string      `json:"difficulty"`
	TimeLimitSec      int         `json:"timeLimitSec"`
	Options           []OptionDTO `json:"options"`
}

type StartResponse struct {
	SessionID uint        `json:"sessionId"`
	ExpiresAt time.Time   `json:"expiresAt"`
	Question  QuestionDTO `json:"question"`
}

type AnswerRequest struct {
	SessionQuestionID uint `json:"sessionQuestionId" binding:"required"`
	OptionID          uint `json:"optionId" binding:"required"`
}

type AnswerResponse struct {
	Correct          bool         `json:"correct"`
	PointsAwarded    int          `json:"pointsAwarded"`
	DifficultyBefore string       `json:"difficultyBefore"`
	DifficultyAfter  string       `json:"difficultyAfter"`
	TotalScore       int          `json:"totalScore"`
	Completed        bool         `json:"completed"`
	Expired          bool         `json:"expired"`
	NextQuestion     *QuestionDTO `json:"nextQuestion,omitempty"`
	Explanation     string `json:"explanation,omitempty"`     // only when assessment allows
	CorrectOptionID *uint  `json:"correctOptionId,omitempty"` // only when assessment allows
	ReportID        *uint  `json:"reportId,omitempty"`        // set when the interview ends
}