package service

import (
	"log"
	"time"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InterviewService struct {
	db      *gorm.DB
	engine  *AdaptiveEngine
	reports *ReportService
}

func NewInterviewService(db *gorm.DB, e *AdaptiveEngine, r *ReportService) *InterviewService {
	return &InterviewService{db: db, engine: e, reports: r}
}

func (s *InterviewService) Start(candidateID, assessmentID uint) (*dto.StartResponse, error) {
	var a models.Assessment
	if err := s.db.Where("id = ? AND active = ? AND assessment_type = ?", assessmentID, true, "MCQ").First(&a).Error; err != nil {
		return nil, NotFound("assessment not found")
	}
	now := time.Now()
	sess := models.InterviewSession{
		CandidateID: candidateID, AssessmentID: a.ID, Status: models.InProgress,
		CurrentDifficulty: models.Medium, // initial difficulty
		StartedAt:         now, ExpiresAt: now.Add(time.Duration(a.DurationMinutes) * time.Minute),
	}
	var first *dto.QuestionDTO
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&sess).Error; err != nil {
			return err
		}
		q, err := s.serveNext(tx, &sess, a.TotalQuestions)
		first = q
		return err
	})
	if err != nil {
		return nil, err
	}
	if first == nil {
		return nil, BadRequest("no questions available for this assessment")
	}
	log.Printf("interview started: session=%d candidate=%d", sess.ID, candidateID)
	return &dto.StartResponse{SessionID: sess.ID, ExpiresAt: sess.ExpiresAt, Question: *first}, nil
}

// Get returns the session state so the UI can resume after a refresh. It also
// enforces expiry on the server.
func (s *InterviewService) Get(candidateID, sessionID uint) (*dto.InterviewState, error) {
	sess, err := s.loadSession(candidateID, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.Status == models.InProgress && time.Now().After(sess.ExpiresAt) {
		if err := s.closeIfOpen(candidateID, sessionID); err != nil {
			return nil, err
		}
		if sess, err = s.loadSession(candidateID, sessionID); err != nil {
			return nil, err
		}
	}
	var a models.Assessment
	if err := s.db.First(&a, sess.AssessmentID).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	remaining := int(sess.ExpiresAt.Sub(now).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	state := &dto.InterviewState{
		SessionID: sess.ID, AssessmentTitle: a.Title, Status: string(sess.Status),
		ExpiresAt: sess.ExpiresAt, ServerTime: now, RemainingSec: remaining,
		AnsweredCount: sess.AnsweredCount, Total: a.TotalQuestions, TotalScore: sess.TotalScore,
	}
	if sess.Status == models.InProgress {
		var sq models.SessionQuestion
		err := s.db.Where("session_id = ? AND answered_at IS NULL", sess.ID).Order("order_no DESC").First(&sq).Error
		if err == nil {
			var q models.Question
			if err := s.db.Preload("Options", orderedOptions).Preload("Topic").First(&q, sq.QuestionID).Error; err != nil {
				return nil, err
			}
			dtoQ := toQuestionDTO(q, sq, a.TotalQuestions)
			state.Question = &dtoQ
		}
		return state, nil
	}
	reportID, err := s.reports.Generate(sess.ID)
	if err != nil {
		return nil, err
	}
	state.ReportID = &reportID
	return state, nil
}

// Answer evaluates a response, updates the session and picks the next question in ONE transaction.
func (s *InterviewService) Answer(candidateID, sessionID uint, req dto.AnswerRequest) (*dto.AnswerResponse, error) {
	var result *dto.AnswerResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		sess, err := lockSession(tx, sessionID, candidateID) // row lock stops double-submits
		if err != nil {
			return err
		}
		if sess.Status != models.InProgress {
			return BadRequest("this interview is already finished")
		}
		now := time.Now()
		if now.After(sess.ExpiresAt) { // server-side timer check
			if err := closeSession(tx, sess, models.Expired, now); err != nil {
				return err
			}
			result = &dto.AnswerResponse{Expired: true, Completed: true, TotalScore: sess.TotalScore}
			return nil // commit the status change
		}

		var sq models.SessionQuestion
		if err := tx.Where("id = ? AND session_id = ?", req.SessionQuestionID, sess.ID).First(&sq).Error; err != nil {
			return NotFound("question not part of this session")
		}
		if sq.AnsweredAt != nil {
			return BadRequest("question already answered") // answers cannot be modified
		}
		var opt models.QuestionOption
		if err := tx.Where("id = ? AND question_id = ?", req.OptionID, sq.QuestionID).First(&opt).Error; err != nil {
			return BadRequest("invalid option for this question")
		}
		var a models.Assessment
		if err := tx.First(&a, sess.AssessmentID).Error; err != nil {
			return err
		}

		correct := opt.IsCorrect
		points := 0
		if correct {
			points = s.engine.Points(sq.Difficulty)
			sess.CorrectCount++
		}
		taken := int(now.Sub(sq.ServedAt).Seconds())
		if err := tx.Create(&models.Response{
			SessionQuestionID: sq.ID, SelectedOptionID: opt.ID, IsCorrect: correct,
			PointsAwarded: points, TimeTakenSec: taken,
		}).Error; err != nil {
			return err
		}
		sq.AnsweredAt = &now
		if err := tx.Save(&sq).Error; err != nil {
			return err
		}

		before := sess.CurrentDifficulty
		next := s.engine.Next(State{
			Difficulty: sess.CurrentDifficulty, ConsecutiveCorrect: sess.ConsecutiveCorrect, ConsecutiveWrong: sess.ConsecutiveWrong,
		}, correct)
		sess.CurrentDifficulty, sess.ConsecutiveCorrect, sess.ConsecutiveWrong = next.Difficulty, next.ConsecutiveCorrect, next.ConsecutiveWrong
		sess.AnsweredCount++
		sess.TotalScore += points

		result = &dto.AnswerResponse{
			Correct: correct, PointsAwarded: points, TotalScore: sess.TotalScore,
			DifficultyBefore: string(before), DifficultyAfter: string(next.Difficulty),
		}
		if a.ShowExplanations { // reveal answer details ONLY if the assessment allows it
			attachExplanation(tx, sq.QuestionID, result)
		}
		if sess.AnsweredCount < a.TotalQuestions {
			q, err := s.serveNext(tx, sess, a.TotalQuestions)
			if err != nil {
				return err
			}
			result.NextQuestion = q
		}
		if result.NextQuestion == nil { // reached the limit or the question bank is exhausted
			result.Completed = true
			log.Printf("interview completed: session=%d score=%d", sess.ID, sess.TotalScore)
			return closeSession(tx, sess, models.Completed, now)
		}
		return tx.Save(sess).Error
	})
	if err != nil {
		return nil, err
	}
	log.Printf("answer submitted: session=%d", sessionID)
	if result.Completed { // report is generated AFTER the answer transaction has committed
		if id, err := s.reports.Generate(sessionID); err != nil {
			log.Printf("report generation failed: session=%d err=%v", sessionID, err)
		} else {
			result.ReportID = &id
		}
	}
	return result, nil
}

// Finish ends the interview early (or after the timer) and returns the report id. Idempotent.
func (s *InterviewService) Finish(candidateID, sessionID uint) (*dto.FinishResponse, error) {
	if err := s.closeIfOpen(candidateID, sessionID); err != nil {
		return nil, err
	}
	reportID, err := s.reports.Generate(sessionID)
	if err != nil {
		return nil, err
	}
	return &dto.FinishResponse{ReportID: reportID}, nil
}

func (s *InterviewService) closeIfOpen(candidateID, sessionID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		sess, err := lockSession(tx, sessionID, candidateID)
		if err != nil {
			return err
		}
		if sess.Status != models.InProgress {
			return nil
		}
		now := time.Now()
		status := models.Completed
		if now.After(sess.ExpiresAt) {
			status = models.Expired
		}
		log.Printf("interview closed: session=%d status=%s", sess.ID, status)
		return closeSession(tx, sess, status, now)
	})
}

func (s *InterviewService) loadSession(candidateID, sessionID uint) (*models.InterviewSession, error) {
	var sess models.InterviewSession
	if err := s.db.Where("id = ? AND candidate_id = ?", sessionID, candidateID).First(&sess).Error; err != nil {
		return nil, NotFound("interview session not found")
	}
	return &sess, nil
}

func lockSession(tx *gorm.DB, sessionID, candidateID uint) (*models.InterviewSession, error) {
	var sess models.InterviewSession
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND candidate_id = ?", sessionID, candidateID).First(&sess).Error
	if err != nil {
		return nil, NotFound("interview session not found")
	}
	return &sess, nil
}

func closeSession(tx *gorm.DB, sess *models.InterviewSession, status models.SessionStatus, at time.Time) error {
	sess.Status, sess.CompletedAt = status, &at
	return tx.Save(sess).Error
}

func attachExplanation(tx *gorm.DB, questionID uint, res *dto.AnswerResponse) {
	var q models.Question
	if err := tx.First(&q, questionID).Error; err == nil {
		res.Explanation = q.Explanation
	}
	var correct models.QuestionOption
	if err := tx.Where("question_id = ? AND is_correct = ?", questionID, true).First(&correct).Error; err == nil {
		id := correct.ID
		res.CorrectOptionID = &id
	}
}

// serveNext picks an unused question (with difficulty fallback) and records it in session_questions.
func (s *InterviewService) serveNext(tx *gorm.DB, sess *models.InterviewSession, total int) (*dto.QuestionDTO, error) {
	for _, d := range s.engine.FallbackOrder(sess.CurrentDifficulty) {
		var qs []models.Question
		used := tx.Model(&models.SessionQuestion{}).Select("question_id").Where("session_id = ?", sess.ID)
		err := tx.Preload("Options", orderedOptions).Preload("Topic").
			Where("active = ? AND question_type = ? AND difficulty = ? AND id NOT IN (?)", true, "MCQ", d, used).
			Order("RAND()").Limit(1).Find(&qs).Error
		if err != nil {
			return nil, err
		}
		if len(qs) == 0 {
			continue
		}
		q := qs[0]
		sq := models.SessionQuestion{
			SessionID: sess.ID, QuestionID: q.ID, OrderNo: sess.AnsweredCount + 1,
			Difficulty: q.Difficulty, ServedAt: time.Now(),
		}
		if err := tx.Create(&sq).Error; err != nil { // UNIQUE(session_id, question_id) is the final guard
			return nil, err
		}
		dtoQ := toQuestionDTO(q, sq, total)
		return &dtoQ, nil
	}
	return nil, nil // nothing left in the bank
}

// toQuestionDTO builds the candidate-facing view. It has no correct flag anywhere.
func toQuestionDTO(q models.Question, sq models.SessionQuestion, total int) dto.QuestionDTO {
	opts := make([]dto.OptionDTO, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, dto.OptionDTO{ID: o.ID, Text: o.OptionText})
	}
	return dto.QuestionDTO{
		SessionQuestionID: sq.ID, Number: sq.OrderNo, Total: total, Text: q.QuestionText,
		Topic: q.Topic.Name, Difficulty: string(q.Difficulty), TimeLimitSec: q.TimeLimitSec, Options: opts,
	}
}