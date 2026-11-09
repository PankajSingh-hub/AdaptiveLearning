package service

import (
	"testing"

	"adaptive-assessment/internal/models"
)

func run(start models.Difficulty, answers ...bool) State {
	e := NewAdaptiveEngine()
	s := State{Difficulty: start}
	for _, a := range answers {
		s = e.Next(s, a)
	}
	return s
}

func TestTwoCorrectIncreases(t *testing.T) {
	if got := run(models.Medium, true, true).Difficulty; got != models.Hard {
		t.Fatalf("want HARD, got %s", got)
	}
}

func TestTwoWrongDecreases(t *testing.T) {
	if got := run(models.Medium, false, false).Difficulty; got != models.Easy {
		t.Fatalf("want EASY, got %s", got)
	}
}

func TestMixedKeepsDifficulty(t *testing.T) {
	if got := run(models.Medium, true, false).Difficulty; got != models.Medium {
		t.Fatalf("want MEDIUM, got %s", got)
	}
}

func TestCannotExceedHard(t *testing.T) {
	if got := run(models.Hard, true, true, true, true).Difficulty; got != models.Hard {
		t.Fatalf("want HARD, got %s", got)
	}
}

func TestCannotGoBelowEasy(t *testing.T) {
	if got := run(models.Easy, false, false, false, false).Difficulty; got != models.Easy {
		t.Fatalf("want EASY, got %s", got)
	}
}

func TestPoints(t *testing.T) {
	e := NewAdaptiveEngine()
	if e.Points(models.Easy) != 1 || e.Points(models.Medium) != 2 || e.Points(models.Hard) != 3 {
		t.Fatal("points mismatch")
	}
}