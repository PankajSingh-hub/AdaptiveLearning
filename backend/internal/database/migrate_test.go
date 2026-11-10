package database

import (
	"os"
	"strings"
	"testing"

	"adaptive-assessment/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Needs an empty MySQL database whose name contains "test". It is SKIPPED unless TEST_DB_DSN
// is set, and every table in that database is dropped first.
func TestMigrationsBuildABlankDatabase(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping MySQL migration test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("cannot connect: %v", err)
	}
	var dbName string
	db.Raw("SELECT DATABASE()").Scan(&dbName)
	if !strings.Contains(dbName, "test") {
		t.Fatalf("refusing to wipe database %q: its name must contain 'test'", dbName)
	}
	wipe(t, db)

	if err := Migrate(dsn); err != nil {
		t.Fatalf("migrating a blank database failed: %v", err)
	}
	if err := Migrate(dsn); err != nil {
		t.Fatalf("running migrations a second time must be a no-op, got: %v", err)
	}

	// Every model must be readable: a missing table or column fails here.
	all := []any{
		&[]models.User{}, &[]models.Topic{}, &[]models.Question{}, &[]models.QuestionOption{},
		&[]models.Assessment{}, &[]models.InterviewSession{}, &[]models.SessionQuestion{},
		&[]models.Response{}, &[]models.InterviewReport{}, &[]models.CodingSession{},
		&[]models.CodingSessionProblem{}, &[]models.CodingSubmission{},
	}
	for _, m := range all {
		if err := db.Limit(1).Find(m).Error; err != nil {
			t.Errorf("%T does not match the migrated schema: %v", m, err)
		}
	}

	// The constraints that GORM's AutoMigrate cannot create.
	for _, idx := range [][2]string{
		{"session_questions", "uq_session_question"},
		{"coding_session_problems", "uq_csp_problem"},
		{"coding_session_problems", "uq_csp_order"},
	} {
		var n int64
		db.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
			idx[0], idx[1]).Scan(&n)
		if n == 0 {
			t.Errorf("missing unique index %s on %s", idx[1], idx[0])
		}
	}

	// Seeding on top of the migrated schema must work end to end.
	Seed(db)
	var questions, codingAssessments int64
	db.Model(&models.Question{}).Count(&questions)
	db.Model(&models.Assessment{}).Where("assessment_type = ?", "CODING").Count(&codingAssessments)
	if questions < 30 || codingAssessments != 1 {
		t.Errorf("seed result wrong: %d questions, %d coding assessments", questions, codingAssessments)
	}
}

func wipe(t *testing.T, db *gorm.DB) {
	t.Helper()
	var names []string
	db.Raw("SELECT table_name AS name FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&names)
	err := db.Connection(func(tx *gorm.DB) error { // one connection, so the session setting applies
		if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
			return err
		}
		for _, n := range names {
			if err := tx.Exec("DROP TABLE IF EXISTS `" + n + "`").Error; err != nil {
				return err
			}
		}
		return tx.Exec("SET FOREIGN_KEY_CHECKS = 1").Error
	})
	if err != nil {
		t.Fatalf("could not empty the test database: %v", err)
	}
}