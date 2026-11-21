package service

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)

func days(offsets ...int) map[string]bool {
	m := map[string]bool{}
	for _, o := range offsets {
		m[testNow.AddDate(0, 0, -o).Format(dayFmt)] = true
	}
	return m
}

func TestStreakCountsConsecutiveDaysEndingToday(t *testing.T) {
	if got := buildStreak(days(0, 1, 2), testNow).Current; got != 3 {
		t.Fatalf("want 3, got %d", got)
	}
}

func TestStreakSurvivesUntilTheDayIsOver(t *testing.T) {
	if got := buildStreak(days(1, 2), testNow).Current; got != 2 {
		t.Fatalf("no activity yet today must not break the streak, got %d", got)
	}
}

func TestStreakBreaksAfterAMissedDay(t *testing.T) {
	if got := buildStreak(days(2, 3), testNow).Current; got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
	if got := buildStreak(days(0, 2, 3), testNow).Current; got != 1 {
		t.Fatalf("gap yesterday: want 1, got %d", got)
	}
}

func TestStreakReturnsSevenDaysOldestFirst(t *testing.T) {
	s := buildStreak(days(0, 3), testNow)
	if len(s.Days) != 7 || !s.Days[6].Today || !s.Days[6].Active || !s.Days[3].Active || s.Days[5].Active {
		t.Fatalf("unexpected days: %+v", s.Days)
	}
}