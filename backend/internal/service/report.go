package service

import (
	"encoding/json"
	"log"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
)

type ReportService struct{ db *gorm.DB }

func NewReportService(db *gorm.DB) *ReportService { return &ReportService{db: db} }

type answerRow struct {
	OrderNo       int
	Difficulty    string
	Topic         string
	IsCorrect     bool
	PointsAwarded int
	TimeTakenSec  int
}

// Generate builds and stores the report for a finished session. It is idempotent:
// calling it again returns the existing report id (session_id is UNIQUE).
func (s *ReportService) Generate(sessionID uint) (uint, error) {
	var existing models.InterviewReport
	if err := s.db.Where("session_id = ?", sessionID).First(&existing).Error; err == nil {
		return existing.ID, nil
	}
	var sess models.InterviewSession
	if err := s.db.First(&sess, sessionID).Error; err != nil {
		return 0, NotFound("interview session not found")
	}
	if sess.Status == models.InProgress {
		return 0, BadRequest("interview is still in progress")
	}
	var assessment models.Assessment
	if err := s.db.First(&assessment, sess.AssessmentID).Error; err != nil {
		return 0, err
	}

	var rows []answerRow
	err := s.db.Table("responses r").
		Select("sq.order_no, sq.difficulty, t.name AS topic, r.is_correct, r.points_awarded, r.time_taken_sec").
		Joins("JOIN session_questions sq ON sq.id = r.session_question_id").
		Joins("JOIN questions q ON q.id = sq.question_id").
		Joins("JOIN topics t ON t.id = q.topic_id").
		Where("sq.session_id = ?", sessionID).Order("sq.order_no").Scan(&rows).Error
	if err != nil {
		return 0, err
	}
	records := make([]AnswerRecord, 0, len(rows))
	for _, r := range rows {
		records = append(records, AnswerRecord{
			OrderNo: r.OrderNo, Difficulty: models.Difficulty(r.Difficulty), Topic: r.Topic,
			Correct: r.IsCorrect, Points: r.PointsAwarded, TimeSec: r.TimeTakenSec,
		})
	}

	a := Analyze(records)
	rep := models.InterviewReport{
		SessionID: sess.ID, CandidateID: sess.CandidateID, AssessmentID: sess.AssessmentID,
		TotalQuestions: assessment.TotalQuestions, Attempted: a.Attempted, CorrectCount: a.Correct,
		Accuracy: a.Accuracy, Score: a.Score, MaxScore: a.MaxScore, AvgResponseSec: a.AvgTimeSec,
		PerformanceLevel: a.Level, StrongTopics: toJSON(a.Strong), WeakTopics: toJSON(a.Weak),
		TopicBreakdown: toJSON(a.Topics), DifficultyBreakdown: toJSON(a.Difficulties),
		DifficultyProgression: toJSON(a.Progression), Recommendations: toJSON(Recommendations(a)),
	}
	if err := s.db.Create(&rep).Error; err != nil {
		// Two requests may race; the UNIQUE key lets only one win, the other reuses it.
		if e2 := s.db.Where("session_id = ?", sessionID).First(&existing).Error; e2 == nil {
			return existing.ID, nil
		}
		return 0, err
	}
	log.Printf("report generated: session=%d report=%d", sessionID, rep.ID)
	return rep.ID, nil
}

// GetByID returns the full report. Candidates can only read their own.
func (s *ReportService) GetByID(id, userID uint, admin bool) (*dto.ReportDTO, error) {
	q := s.db
	if !admin {
		q = q.Where("candidate_id = ?", userID)
	}
	var rep models.InterviewReport
	if err := q.First(&rep, id).Error; err != nil {
		return nil, NotFound("report not found")
	}
	var user models.User
	var assessment models.Assessment
	var sess models.InterviewSession
	s.db.First(&user, rep.CandidateID)
	s.db.First(&assessment, rep.AssessmentID)
	s.db.First(&sess, rep.SessionID)
	date := sess.StartedAt
	if sess.CompletedAt != nil {
		date = *sess.CompletedAt
	}
	return &dto.ReportDTO{
		Warnings: sess.Warnings, Terminated: sess.Terminated,
		ID: rep.ID, SessionID: rep.SessionID, CandidateID: rep.CandidateID,
		CandidateName: user.Name, AssessmentTitle: assessment.Title, Date: date,
		TotalQuestions: rep.TotalQuestions, Attempted: rep.Attempted, Correct: rep.CorrectCount,
		Incorrect: rep.Attempted - rep.CorrectCount, Accuracy: rep.Accuracy,
		Score: rep.Score, MaxScore: rep.MaxScore, AvgResponseSec: rep.AvgResponseSec,
		PerformanceLevel:      rep.PerformanceLevel,
		StrongTopics:          fromJSON[[]string](rep.StrongTopics),
		WeakTopics:            fromJSON[[]string](rep.WeakTopics),
		TopicBreakdown:        fromJSON[[]dto.TopicStatDTO](rep.TopicBreakdown),
		DifficultyBreakdown:   fromJSON[[]dto.DifficultyStatDTO](rep.DifficultyBreakdown),
		DifficultyProgression: fromJSON[[]string](rep.DifficultyProgression),
		Recommendations:       fromJSON[[]string](rep.Recommendations),
	}, nil
}

// List returns report summaries. Non-admins always see only their own reports.
func (s *ReportService) List(userID uint, admin bool, candidateID uint, page, size int) (*dto.PagedReports, error) {
	if page < 0 {
		page = 0
	}
	if size < 1 || size > 100 {
		size = 10
	}
	base := s.db.Table("interview_reports r")
	if !admin {
		base = base.Where("r.candidate_id = ?", userID)
	} else if candidateID > 0 {
		base = base.Where("r.candidate_id = ?", candidateID)
	}
	base = base.Session(&gorm.Session{})

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	items := []dto.ReportSummary{}
	err := base.Select("r.id, r.session_id, r.candidate_id, u.name AS candidate_name, a.title AS assessment_title, "+
		"r.score, r.max_score, r.accuracy, r.performance_level AS level, r.created_at").
		Joins("JOIN users u ON u.id = r.candidate_id").
		Joins("JOIN assessments a ON a.id = r.assessment_id").
		Order("r.created_at DESC, r.id DESC").Offset(page * size).Limit(size).Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return &dto.PagedReports{Items: items, Page: page, Size: size, TotalItems: total, TotalPages: totalPages(total, size)}, nil
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func fromJSON[T any](raw string) T {
	var v T
	_ = json.Unmarshal([]byte(raw), &v)
	return v
}