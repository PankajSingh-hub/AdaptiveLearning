package service

import "adaptive-assessment/internal/models"

// State is everything the engine needs. It has no DB dependency, which makes it unit-testable.
type State struct {
	Difficulty         models.Difficulty
	ConsecutiveCorrect int
	ConsecutiveWrong   int
}

const streakThreshold = 2

// AdaptiveEngine decides difficulty. It is intentionally stateless.
type AdaptiveEngine struct{}

func NewAdaptiveEngine() *AdaptiveEngine { return &AdaptiveEngine{} }

// Next applies the rule: 2 correct in a row -> harder, 2 wrong in a row -> easier, else unchanged.
// Streaks reset after a level change so the next level change needs fresh evidence.
func (e *AdaptiveEngine) Next(s State, correct bool) State {
	if correct {
		s.ConsecutiveCorrect++
		s.ConsecutiveWrong = 0
		if s.ConsecutiveCorrect >= streakThreshold {
			s.Difficulty = harder(s.Difficulty)
			s.ConsecutiveCorrect = 0
		}
	} else {
		s.ConsecutiveWrong++
		s.ConsecutiveCorrect = 0
		if s.ConsecutiveWrong >= streakThreshold {
			s.Difficulty = easier(s.Difficulty)
			s.ConsecutiveWrong = 0
		}
	}
	return s
}

// Points are difficulty-weighted: Easy=1, Medium=2, Hard=3.
func (e *AdaptiveEngine) Points(d models.Difficulty) int {
	switch d {
	case models.Hard:
		return 3
	case models.Medium:
		return 2
	default:
		return 1
	}
}

// FallbackOrder is the difficulty search order when the bank has no unused question at the target level.
func (e *AdaptiveEngine) FallbackOrder(target models.Difficulty) []models.Difficulty {
	switch target {
	case models.Hard:
		return []models.Difficulty{models.Hard, models.Medium, models.Easy}
	case models.Easy:
		return []models.Difficulty{models.Easy, models.Medium, models.Hard}
	default:
		return []models.Difficulty{models.Medium, models.Easy, models.Hard}
	}
}

func harder(d models.Difficulty) models.Difficulty {
	if d == models.Easy {
		return models.Medium
	}
	return models.Hard // HARD stays HARD
}

func easier(d models.Difficulty) models.Difficulty {
	if d == models.Hard {
		return models.Medium
	}
	return models.Easy // EASY stays EASY
}