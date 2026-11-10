package dto

import "time"

type TopicStatDTO struct {
	Topic     string  `json:"topic"`
	Attempted int     `json:"attempted"`
	Correct   int     `json:"correct"`
	Accuracy  float64 `json:"accuracy"`
}

type DifficultyStatDTO struct {
	Difficulty string  `json:"difficulty"`
	Attempted  int     `json:"attempted"`
	Correct    int     `json:"correct"`
	Accuracy   float64 `json:"accuracy"`
}

type ReportDTO struct {
	ID                    uint                `json:"id"`
	SessionID             uint                `json:"sessionId"`
	CandidateID           uint                `json:"candidateId"`
	CandidateName         string              `json:"candidateName"`
	AssessmentTitle       string              `json:"assessmentTitle"`
	Date                  time.Time           `json:"date"`
	TotalQuestions        int                 `json:"totalQuestions"`
	Attempted             int                 `json:"attempted"`
	Correct               int                 `json:"correct"`
	Incorrect             int                 `json:"incorrect"`
	Accuracy              float64             `json:"accuracy"`
	Score                 int                 `json:"score"`
	MaxScore              int                 `json:"maxScore"`
	AvgResponseSec        float64             `json:"avgResponseSec"`
	PerformanceLevel      string              `json:"performanceLevel"`
	StrongTopics          []string            `json:"strongTopics"`
	WeakTopics            []string            `json:"weakTopics"`
	TopicBreakdown        []TopicStatDTO      `json:"topicBreakdown"`
	DifficultyBreakdown   []DifficultyStatDTO `json:"difficultyBreakdown"`
	DifficultyProgression []string            `json:"difficultyProgression"`
	Recommendations       []string            `json:"recommendations"`
	Warnings   int  `json:"warnings"`
Terminated bool `json:"terminated"`
}

type ReportSummary struct {
	ID              uint      `json:"id"`
	SessionID       uint      `json:"sessionId"`
	CandidateID     uint      `json:"candidateId"`
	CandidateName   string    `json:"candidateName"`
	AssessmentTitle string    `json:"assessmentTitle"`
	Score           int       `json:"score"`
	MaxScore        int       `json:"maxScore"`
	Accuracy        float64   `json:"accuracy"`
	Level           string    `json:"level"`
	CreatedAt       time.Time `json:"createdAt"`
}

type PagedReports struct {
	Items      []ReportSummary `json:"items"`
	Page       int             `json:"page"`
	Size       int             `json:"size"`
	TotalItems int64           `json:"totalItems"`
	TotalPages int             `json:"totalPages"`
}

// InterviewState lets the frontend resume a session after a page refresh.
type InterviewState struct {
	SessionID       uint         `json:"sessionId"`
	AssessmentTitle string       `json:"assessmentTitle"`
	Status          string       `json:"status"`
	ExpiresAt       time.Time    `json:"expiresAt"`
	ServerTime      time.Time    `json:"serverTime"`
	RemainingSec    int          `json:"remainingSec"`
	AnsweredCount   int          `json:"answeredCount"`
	Total           int          `json:"total"`
	TotalScore      int          `json:"totalScore"`
	Question        *QuestionDTO `json:"question,omitempty"`
	ReportID        *uint        `json:"reportId,omitempty"`
	Warnings   int  `json:"warnings"`

	
}

type FinishResponse struct {
	ReportID uint `json:"reportId"`
}

// ---- Analytics ----

type TrendPoint struct {
	TakenAt  time.Time `json:"date"`
	Accuracy float64   `json:"accuracy"`
	Score    int       `json:"score"`
}

type CandidateAnalytics struct {
	CandidateID           uint                `json:"candidateId"`
	CandidateName         string              `json:"candidateName"`
	CompletedInterviews   int                 `json:"completedInterviews"`
	QuestionsAttempted    int                 `json:"questionsAttempted"`
	CorrectAnswers        int                 `json:"correctAnswers"`
	IncorrectAnswers      int                 `json:"incorrectAnswers"`
	Accuracy              float64             `json:"accuracy"`
	TotalScore            int                 `json:"totalScore"`
	AvgScore              float64             `json:"avgScore"`
	AvgResponseSec        float64             `json:"avgResponseSec"`
	TopicPerformance      []TopicStatDTO      `json:"topicPerformance"`
	DifficultyPerformance []DifficultyStatDTO `json:"difficultyPerformance"`
	Trend                 []TrendPoint        `json:"trend"`
	RecommendedTopics     []string            `json:"recommendedTopics"`
	RecentReports         []ReportSummary     `json:"recentReports"`
}

type LabelCount struct {
	Label string `json:"label"`
	Total int64  `json:"count"`
}

type TopCandidate struct {
	CandidateID uint    `json:"candidateId"`
	Name        string  `json:"name"`
	Interviews  int     `json:"interviews"`
	AvgAccuracy float64 `json:"avgAccuracy"`
	AvgScore    float64 `json:"avgScore"`
}

type AdminAnalytics struct {
	TotalCandidates        int64           `json:"totalCandidates"`
	TotalAssessments       int64           `json:"totalAssessments"`
	TotalQuestions         int64           `json:"totalQuestions"`
	CompletedInterviews    int64           `json:"completedInterviews"`
	AvgScore               float64         `json:"avgScore"`
	AvgAccuracy            float64         `json:"avgAccuracy"`
	DifficultyDistribution []LabelCount    `json:"difficultyDistribution"`
	TopicDistribution      []LabelCount    `json:"topicDistribution"`
	RecentInterviews       []ReportSummary `json:"recentInterviews"`
	TopCandidates          []TopCandidate  `json:"topCandidates"`
}

// ---- Candidates (admin) ----

type CandidateListItem struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	JoinedAt      time.Time  `json:"joinedAt"`
	Interviews    int        `json:"interviews"`
	AvgAccuracy   float64    `json:"avgAccuracy"`
	LastInterview *time.Time `json:"lastInterview"`
}

type PagedCandidates struct {
	Items      []CandidateListItem `json:"items"`
	Page       int                 `json:"page"`
	Size       int                 `json:"size"`
	TotalItems int64               `json:"totalItems"`
	TotalPages int                 `json:"totalPages"`
}

type CandidateDetail struct {
	ID       uint            `json:"id"`
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	JoinedAt time.Time       `json:"joinedAt"`
	Reports  []ReportSummary `json:"reports"`
}