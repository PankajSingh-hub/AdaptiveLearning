package service

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// These tests need a MySQL database whose name contains "test". They are SKIPPED
// unless TEST_DB_DSN is set, and the tables in that database are dropped and recreated.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping MySQL integration test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("cannot connect to test database: %v", err)
	}
	var name string
	db.Raw("SELECT DATABASE()").Scan(&name)
	if !strings.Contains(name, "test") {
		t.Fatalf("refusing to wipe database %q: its name must contain 'test'", name)
	}
	tables := []interface{}{ // dependency order; DropTable removes them in reverse
		&models.User{}, &models.Topic{}, &models.Assessment{}, &models.Question{}, &models.QuestionOption{},
		&models.InterviewSession{}, &models.SessionQuestion{}, &models.Response{}, &models.InterviewReport{},
	}
	if err := db.Migrator().DropTable(tables...); err != nil {
		t.Fatalf("drop tables: %v", err)
	}
	if err := db.AutoMigrate(tables...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

type ivFixture struct {
	db           *gorm.DB
	svc          *InterviewService
	candidateID  uint
	assessmentID uint
}

func mustCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("test setup failed: %v", err)
	}
}

// newIvFixture creates perLevel questions at each difficulty and an assessment of `total` questions.
// In every question the FIRST option is the correct one.
func newIvFixture(t *testing.T, perLevel, total int) ivFixture {
	t.Helper()
	db := openTestDB(t)
	user := models.User{Name: "Test Candidate", Email: "tc@test.com", PasswordHash: "x", Role: models.RoleCandidate}
	mustCreate(t, db, &user)
	topic := models.Topic{Name: "Data Structures"}
	mustCreate(t, db, &topic)
	for _, d := range []models.Difficulty{models.Easy, models.Medium, models.Hard} {
		for i := 1; i <= perLevel; i++ {
			mustCreate(t, db, &models.Question{
				QuestionText: fmt.Sprintf("%s question number %d", d, i), TopicID: topic.ID, Category: "DSA",
				Difficulty: d, QuestionType: "MCQ", TimeLimitSec: 60, Active: true,
				Options: []models.QuestionOption{{OptionText: "right", IsCorrect: true}, {OptionText: "wrong"}},
			})
		}
	}
	a := models.Assessment{Title: "Test assessment", TotalQuestions: total, DurationMinutes: 30, Active: true}
	mustCreate(t, db, &a)
	reports := NewReportService(db)
	return ivFixture{db: db, svc: NewInterviewService(db, NewAdaptiveEngine(), reports), candidateID: user.ID, assessmentID: a.ID}
}

func (f ivFixture) answer(t *testing.T, sessionID uint, q *dto.QuestionDTO, correct bool) *dto.AnswerResponse {
	t.Helper()
	opt := q.Options[1]
	if correct {
		opt = q.Options[0]
	}
	res, err := f.svc.Answer(f.candidateID, sessionID, dto.AnswerRequest{SessionQuestionID: q.SessionQuestionID, OptionID: opt.ID})
	if err != nil {
		t.Fatalf("answer failed: %v", err)
	}
	return res
}

func TestNoDuplicateQuestionsInOneSession(t *testing.T) {
	f := newIvFixture(t, 4, 10) // 12 questions in the bank, 10 needed
	start, err := f.svc.Start(f.candidateID, f.assessmentID)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	q := &start.Question
	for q != nil {
		if seen[q.Text] {
			t.Fatalf("question served twice: %q", q.Text)
		}
		seen[q.Text] = true
		q = f.answer(t, start.SessionID, q, true).NextQuestion
	}
	if len(seen) != 10 {
		t.Fatalf("expected 10 distinct questions, got %d", len(seen))
	}
}

func TestDifficultyFollowsAnswersThroughTheService(t *testing.T) {
	f := newIvFixture(t, 3, 6)
	start, err := f.svc.Start(f.candidateID, f.assessmentID)
	if err != nil {
		t.Fatal(err)
	}
	q := &start.Question
	if q.Difficulty != "MEDIUM" {
		t.Fatalf("first question should be MEDIUM, got %s", q.Difficulty)
	}
	r1 := f.answer(t, start.SessionID, q, true)
	if r1.DifficultyAfter != "MEDIUM" {
		t.Fatalf("after 1 correct want MEDIUM, got %s", r1.DifficultyAfter)
	}
	r2 := f.answer(t, start.SessionID, r1.NextQuestion, true)
	if r2.DifficultyAfter != "HARD" || r2.NextQuestion.Difficulty != "HARD" {
		t.Fatalf("after 2 correct want HARD, got %s / %s", r2.DifficultyAfter, r2.NextQuestion.Difficulty)
	}
	r3 := f.answer(t, start.SessionID, r2.NextQuestion, false)
	if r3.DifficultyAfter != "HARD" {
		t.Fatalf("after 1 wrong want HARD, got %s", r3.DifficultyAfter)
	}
	r4 := f.answer(t, start.SessionID, r3.NextQuestion, false)
	if r4.DifficultyAfter != "MEDIUM" {
		t.Fatalf("after 2 wrong want MEDIUM, got %s", r4.DifficultyAfter)
	}
}

func TestInterviewEndsAndReportIsGeneratedWhenBankRunsOut(t *testing.T) {
	f := newIvFixture(t, 1, 10) // only 3 questions exist but 10 are requested
	start, err := f.svc.Start(f.candidateID, f.assessmentID)
	if err != nil {
		t.Fatal(err)
	}
	q := &start.Question
	var last *dto.AnswerResponse
	for i := 0; i < 5 && q != nil; i++ { // safety cap against an endless loop
		last = f.answer(t, start.SessionID, q, true)
		q = last.NextQuestion
	}
	if last == nil || !last.Completed || last.ReportID == nil {
		t.Fatalf("interview should complete with a report: %+v", last)
	}
	again, err := f.svc.reports.Generate(start.SessionID)
	if err != nil || again != *last.ReportID {
		t.Fatalf("Generate must be idempotent: %d vs %d (%v)", again, *last.ReportID, err)
	}
}

func TestAnswersCannotBeSubmittedTwice(t *testing.T) {
	f := newIvFixture(t, 3, 5)
	start, _ := f.svc.Start(f.candidateID, f.assessmentID)
	q := start.Question
	f.answer(t, start.SessionID, &q, true)
	_, err := f.svc.Answer(f.candidateID, start.SessionID, dto.AnswerRequest{SessionQuestionID: q.SessionQuestionID, OptionID: q.Options[1].ID})
	if err == nil {
		t.Fatal("answering the same question twice must fail")
	}
}

func TestExpiredSessionIsClosedByTheServer(t *testing.T) {
	f := newIvFixture(t, 3, 5)
	start, _ := f.svc.Start(f.candidateID, f.assessmentID)
	f.db.Model(&models.InterviewSession{}).Where("id = ?", start.SessionID).Update("expires_at", time.Now().Add(-time.Minute))
	q := start.Question
	res, err := f.svc.Answer(f.candidateID, start.SessionID, dto.AnswerRequest{SessionQuestionID: q.SessionQuestionID, OptionID: q.Options[0].ID})
	if err != nil || !res.Expired {
		t.Fatalf("expired session must report Expired (err=%v res=%+v)", err, res)
	}
	var sess models.InterviewSession
	f.db.First(&sess, start.SessionID)
	if sess.Status != models.Expired {
		t.Fatalf("session status should be EXPIRED, got %s", sess.Status)
	}
}

func TestCandidateCannotAnswerSomeoneElsesSession(t *testing.T) {
	f := newIvFixture(t, 3, 5)
	start, _ := f.svc.Start(f.candidateID, f.assessmentID)
	q := start.Question
	_, err := f.svc.Answer(f.candidateID+999, start.SessionID, dto.AnswerRequest{SessionQuestionID: q.SessionQuestionID, OptionID: q.Options[0].ID})
	if err == nil {
		t.Fatal("another candidate must not be able to answer this session")
	}
}