package service

import (
	"log"
	"math/rand"
	"time"

	"adaptive-assessment/internal/coding"
	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxFailures  = 5
	maxCodeBytes = 60000
)

var basePoints = map[string]int{"EASY": 10, "MEDIUM": 20, "HARD": 30}

type CodingService struct {
	db *gorm.DB
	ex coding.Executor
}

func NewCodingService(db *gorm.DB, ex coding.Executor) *CodingService {
	return &CodingService{db: db, ex: ex}
}

func (s *CodingService) Languages() []dto.CodingLanguage {
	out := []dto.CodingLanguage{}
	for _, l := range s.ex.Languages() {
		out = append(out, dto.CodingLanguage{ID: l.ID, Label: l.Label})
	}
	return out
}

// Start picks a random subset of the problem catalog. Resuming is automatic:
// an unexpired session for the same assessment is returned instead of creating a new one.
func (s *CodingService) Start(candidateID, assessmentID uint) (*dto.CodingStartResponse, error) {
	if len(s.ex.Languages()) == 0 {
		return nil, BadRequest("no programming language toolchain is installed on the server")
	}
	var a models.Assessment
	if err := s.db.Where("id = ? AND active = ? AND assessment_type = ?", assessmentID, true, "CODING").First(&a).Error; err != nil {
		return nil, NotFound("coding assessment not found")
	}
	var open models.CodingSession
	err := s.db.Where("candidate_id = ? AND assessment_id = ? AND status = ? AND expires_at > ?",
		candidateID, a.ID, models.InProgress, time.Now()).First(&open).Error
	if err == nil {
		return &dto.CodingStartResponse{SessionID: open.ID}, nil
	}

	n := min(a.TotalQuestions, len(coding.Catalog))
	picked := rand.Perm(len(coding.Catalog))[:n]
	now := time.Now()
	sess := models.CodingSession{
		CandidateID: candidateID, AssessmentID: a.ID, Status: models.InProgress,
		StartedAt: now, ExpiresAt: now.Add(time.Duration(a.DurationMinutes) * time.Minute),
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&sess).Error; err != nil {
			return err
		}
		rows := make([]models.CodingSessionProblem, 0, n)
		for i, idx := range picked {
			rows = append(rows, models.CodingSessionProblem{
				SessionID: sess.ID, ProblemSlug: coding.Catalog[idx].Slug, OrderNo: i + 1, Status: "OPEN",
			})
		}
		return tx.Create(&rows).Error
	})
	if err != nil {
		return nil, err
	}
	log.Printf("coding session started: session=%d candidate=%d problems=%d", sess.ID, candidateID, n)
	return &dto.CodingStartResponse{SessionID: sess.ID}, nil
}

func (s *CodingService) GetSession(candidateID, id uint) (*dto.CodingSessionState, error) {
	sess, err := s.loadSession(candidateID, id)
	if err != nil {
		return nil, err
	}
	var a models.Assessment
	s.db.First(&a, sess.AssessmentID)
	var rows []models.CodingSessionProblem
	if err := s.db.Where("session_id = ?", id).Order("order_no").Find(&rows).Error; err != nil {
		return nil, err
	}
	state := &dto.CodingSessionState{
		SessionID: sess.ID, AssessmentTitle: a.Title, Status: string(sess.Status), Score: sess.Score,
		Problems: make([]dto.CodingProblemSummary, 0, len(rows)),
	}
	if sess.Status == models.InProgress {
		state.RemainingSec = max(0, int(time.Until(sess.ExpiresAt).Seconds()))
	}
	for _, r := range rows {
		p, ok := coding.Find(r.ProblemSlug)
		if !ok {
			continue
		}
		state.MaxScore += basePoints[p.Difficulty]
		if state.Current == 0 && r.Status == "OPEN" && sess.Status == models.InProgress {
			state.Current = r.OrderNo
		}
		state.Problems = append(state.Problems, dto.CodingProblemSummary{
			Order: r.OrderNo, Slug: p.Slug, Title: p.Title, Difficulty: p.Difficulty, Status: r.Status,
			Failures: r.Failures, Points: r.Points, RuntimeMs: r.RuntimeMs, Language: r.Language, TimeOpt: p.TimeOpt,
		})
	}
	return state, nil
}

func (s *CodingService) GetProblem(candidateID, id uint, order int) (*dto.CodingProblemView, error) {
	sess, err := s.loadSession(candidateID, id)
	if err != nil {
		return nil, err
	}
	row, p, err := s.loadProblem(sess.ID, order)
	if err != nil {
		return nil, err
	}
	view := &dto.CodingProblemView{
		Order: row.OrderNo, Slug: p.Slug, Title: p.Title, Difficulty: p.Difficulty, Description: p.Description,
		Starter: map[string]string{}, Hints: []string{}, Failures: row.Failures, MaxFailures: maxFailures, Status: row.Status,
	}
	for _, l := range s.ex.Languages() {
		view.Starter[l.ID] = coding.Starter(p, l.ID)
	}
	for i := 0; i < coding.SampleCount; i++ {
		t := p.Tests[i]
		view.Examples = append(view.Examples, dto.CodingExample{
			Input: coding.Render(p, t.Args), Output: coding.RenderOutput(p.Ref(t.Args)),
		})
	}
	for i := 0; i < min(row.Failures, maxFailures); i++ {
		view.Hints = append(view.Hints, p.Hints[i])
	}
	return view, nil
}

// Run tests the visible examples only. It never changes the database or uses a hint.
func (s *CodingService) Run(candidateID, id uint, order int, req dto.CodeRequest) (*dto.CodingResult, error) {
	_, _, p, err := s.prepare(candidateID, id, order, req)
	if err != nil {
		return nil, err
	}
	out := coding.Judge(s.ex, p, req.Language, req.Code, coding.ModeRun)
	res := toCodingResult(out, "run")
	res.Analysis = coding.Analyze(p, out)
	return res, nil
}

func (s *CodingService) Submit(candidateID, id uint, order int, req dto.CodeRequest) (*dto.CodingResult, error) {
	sess, row, p, err := s.prepare(candidateID, id, order, req)
	if err != nil {
		return nil, err
	}
	// Judge OUTSIDE the transaction: running code can take seconds and must not hold row locks.
	out := coding.Judge(s.ex, p, req.Language, req.Code, coding.ModeSubmit)

	var res *dto.CodingResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var cur models.CodingSessionProblem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cur, row.ID).Error; err != nil {
			return NotFound("problem not found")
		}
		if cur.Status != "OPEN" {
			return BadRequest("this problem is already finished")
		}
		res = toCodingResult(out, "submit")
		res.Analysis = coding.Analyze(p, out)
		cur.Attempts++
		cur.Language = req.Language

		switch out.Verdict {
		case coding.CompileFail:
			// compile errors never use up a hint
		case coding.Accepted:
			points := basePoints[p.Difficulty] * (10 - cur.Failures) / 10
			now := time.Now()
			cur.Status, cur.Points, cur.RuntimeMs, cur.SolvedAt = "SOLVED", points, out.MaxMillis, &now
			res.Solved, res.Points = true, points
			fb := coding.BuildFeedback(p, out, cur.Failures)
			res.Feedback = &dto.CodingFeedback{TimeOptimal: fb.TimeOptimal, SpaceOptimal: fb.SpaceOptimal,
				Approach: fb.Approach, Notes: fb.Notes, RuntimeMs: fb.RuntimeMs}
			if err := tx.Model(&models.CodingSession{}).Where("id = ?", sess.ID).
				UpdateColumn("score", gorm.Expr("score + ?", points)).Error; err != nil {
				return err
			}
		default:
			cur.Failures++
			res.HintNumber, res.Hint = cur.Failures, p.Hints[cur.Failures-1]
			if cur.Failures >= maxFailures {
				cur.Status = "FAILED"
				res.Exhausted = true
			}
		}
		res.Failures = cur.Failures

		if err := tx.Save(&cur).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.CodingSubmission{
			SessionProblemID: cur.ID, Language: req.Language, Code: req.Code, Verdict: string(out.Verdict),
			Passed: out.Passed, Total: out.Total, RuntimeMs: out.MaxMillis,
		}).Error; err != nil {
			return err
		}
		if cur.Status == "OPEN" {
			return nil
		}
		var next models.CodingSessionProblem
		err := tx.Where("session_id = ? AND status = ?", sess.ID, "OPEN").Order("order_no").First(&next).Error
		if err == nil {
			res.NextOrder = next.OrderNo
			return nil
		}
		res.SessionDone = true
		return closeCoding(tx, sess.ID, models.Completed)
	})
	if err != nil {
		return nil, err
	}
	log.Printf("coding submission: session=%d problem=%s verdict=%s", id, p.Slug, out.Verdict)
	return res, nil
}

func (s *CodingService) Finish(candidateID, id uint) error {
	sess, err := s.loadSession(candidateID, id)
	if err != nil {
		return err
	}
	if sess.Status != models.InProgress {
		return nil
	}
	return s.db.Transaction(func(tx *gorm.DB) error { return closeCoding(tx, id, models.Completed) })
}

// ---- helpers ----

func (s *CodingService) prepare(candidateID, id uint, order int, req dto.CodeRequest) (*models.CodingSession, *models.CodingSessionProblem, coding.Problem, error) {
	sess, err := s.loadSession(candidateID, id)
	if err != nil {
		return nil, nil, coding.Problem{}, err
	}
	if sess.Status != models.InProgress {
		return nil, nil, coding.Problem{}, BadRequest("this coding session is already finished")
	}
	row, p, err := s.loadProblem(sess.ID, order)
	if err != nil {
		return nil, nil, p, err
	}
	if row.Status != "OPEN" {
		return nil, nil, p, BadRequest("this problem is already finished")
	}
	if len(req.Code) > maxCodeBytes {
		return nil, nil, p, BadRequest("code is too long")
	}
	for _, l := range s.ex.Languages() {
		if l.ID == req.Language {
			return sess, row, p, nil
		}
	}
	return nil, nil, p, BadRequest("that language is not available on this server")
}

func (s *CodingService) loadSession(candidateID, id uint) (*models.CodingSession, error) {
	var sess models.CodingSession
	if err := s.db.Where("id = ? AND candidate_id = ?", id, candidateID).First(&sess).Error; err != nil {
		return nil, NotFound("coding session not found")
	}
	if sess.Status == models.InProgress && time.Now().After(sess.ExpiresAt) {
		if err := s.db.Transaction(func(tx *gorm.DB) error { return closeCoding(tx, sess.ID, models.Expired) }); err != nil {
			return nil, err
		}
		s.db.First(&sess, sess.ID)
	}
	return &sess, nil
}

func (s *CodingService) loadProblem(sessionID uint, order int) (*models.CodingSessionProblem, coding.Problem, error) {
	var row models.CodingSessionProblem
	if err := s.db.Where("session_id = ? AND order_no = ?", sessionID, order).First(&row).Error; err != nil {
		return nil, coding.Problem{}, NotFound("problem not found")
	}
	p, ok := coding.Find(row.ProblemSlug)
	if !ok {
		return nil, p, NotFound("problem is no longer available")
	}
	return &row, p, nil
}

func closeCoding(tx *gorm.DB, sessionID uint, status models.SessionStatus) error {
	now := time.Now()
	if err := tx.Model(&models.CodingSession{}).Where("id = ?", sessionID).
		Updates(map[string]any{"status": status, "completed_at": now}).Error; err != nil {
		return err
	}
	return tx.Model(&models.CodingSessionProblem{}).Where("session_id = ? AND status = ?", sessionID, "OPEN").
		Update("status", "SKIPPED").Error
}

func toCodingResult(o coding.Outcome, mode string) *dto.CodingResult {
	cases := make([]dto.CodingCase, 0, len(o.Cases))
	for _, c := range o.Cases {
		cases = append(cases, dto.CodingCase{Tag: c.Tag, Visible: c.Visible, Passed: c.Passed, Status: c.Status,
			Input: c.Input, Expected: c.Expected, Got: c.Got, TimeMs: c.Millis})
	}
	return &dto.CodingResult{Mode: mode, Verdict: string(o.Verdict), Passed: o.Passed, Total: o.Total,
		CompileOutput: o.CompileOutput, Detail: o.Detail, Cases: cases}
}