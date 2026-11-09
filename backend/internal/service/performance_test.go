package service

import (
	"testing"

	"adaptive-assessment/internal/models"
)

func TestAnalyzeScoresTopicsAndLevel(t *testing.T) {
	a := Analyze([]AnswerRecord{
		{1, models.Medium, "DBMS", true, 2, 10},
		{2, models.Medium, "DBMS", true, 2, 20},
		{3, models.Hard, "Operating Systems", false, 0, 30},
	})
	if a.Attempted != 3 || a.Correct != 2 {
		t.Fatalf("counts wrong: %+v", a)
	}
	if a.Score != 4 || a.MaxScore != 7 {
		t.Fatalf("score wrong: %d/%d", a.Score, a.MaxScore)
	}
	if len(a.Strong) != 1 || a.Strong[0] != "DBMS" {
		t.Fatalf("strong topics wrong: %v", a.Strong)
	}
	if len(a.Weak) != 1 || a.Weak[0] != "Operating Systems" {
		t.Fatalf("weak topics wrong: %v", a.Weak)
	}
	if a.AvgTimeSec != 20 {
		t.Fatalf("avg time wrong: %v", a.AvgTimeSec)
	}
	if a.Level != "AVERAGE" { // 66.7% accuracy
		t.Fatalf("level wrong: %s", a.Level)
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	a := Analyze(nil)
	if a.Level != "INCOMPLETE" || a.Accuracy != 0 {
		t.Fatalf("empty analysis wrong: %+v", a)
	}
}