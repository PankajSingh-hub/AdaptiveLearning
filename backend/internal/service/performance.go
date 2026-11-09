package service

import (
	"math"
	"sort"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"
)

// AnswerRecord is one answered question. The analyzer takes plain data and has
// no DB dependency, so it is easy to unit test.
type AnswerRecord struct {
	OrderNo    int
	Difficulty models.Difficulty
	Topic      string
	Correct    bool
	Points     int
	TimeSec    int
}

type Analysis struct {
	Attempted    int
	Correct      int
	Accuracy     float64
	Score        int
	MaxScore     int
	AvgTimeSec   float64
	Topics       []dto.TopicStatDTO
	Difficulties []dto.DifficultyStatDTO
	Progression  []string
	Strong       []string
	Weak         []string
	Level        string
}

const (
	strongThreshold = 75.0 // topic accuracy >= 75% counts as strong
	weakThreshold   = 60.0 // topic accuracy <  60% counts as weak
)

var scoring = NewAdaptiveEngine() // single source of truth for points per difficulty

func Analyze(records []AnswerRecord) Analysis {
	a := Analysis{
		Topics: []dto.TopicStatDTO{}, Difficulties: []dto.DifficultyStatDTO{},
		Progression: []string{}, Strong: []string{}, Weak: []string{},
	}
	topics := map[string]*dto.TopicStatDTO{}
	diffs := map[models.Difficulty]*dto.DifficultyStatDTO{}
	totalTime := 0

	for _, r := range records {
		a.Attempted++
		totalTime += r.TimeSec
		a.Progression = append(a.Progression, string(r.Difficulty))
		a.MaxScore += scoring.Points(r.Difficulty)
		if r.Correct {
			a.Correct++
			a.Score += r.Points
		}
		t := topics[r.Topic]
		if t == nil {
			t = &dto.TopicStatDTO{Topic: r.Topic}
			topics[r.Topic] = t
		}
		d := diffs[r.Difficulty]
		if d == nil {
			d = &dto.DifficultyStatDTO{Difficulty: string(r.Difficulty)}
			diffs[r.Difficulty] = d
		}
		t.Attempted++
		d.Attempted++
		if r.Correct {
			t.Correct++
			d.Correct++
		}
	}

	a.Accuracy = pct(a.Correct, a.Attempted)
	if a.Attempted > 0 {
		a.AvgTimeSec = round1(float64(totalTime) / float64(a.Attempted))
	}
	for _, t := range topics {
		t.Accuracy = pct(t.Correct, t.Attempted)
		a.Topics = append(a.Topics, *t)
	}
	sort.Slice(a.Topics, func(i, j int) bool {
		if a.Topics[i].Accuracy != a.Topics[j].Accuracy {
			return a.Topics[i].Accuracy > a.Topics[j].Accuracy
		}
		return a.Topics[i].Topic < a.Topics[j].Topic
	})
	for _, t := range a.Topics {
		switch {
		case t.Accuracy >= strongThreshold:
			a.Strong = append(a.Strong, t.Topic)
		case t.Accuracy < weakThreshold:
			a.Weak = append(a.Weak, t.Topic)
		}
	}
	for _, d := range []models.Difficulty{models.Easy, models.Medium, models.Hard} {
		if s, ok := diffs[d]; ok {
			s.Accuracy = pct(s.Correct, s.Attempted)
			a.Difficulties = append(a.Difficulties, *s)
		}
	}
	a.Level = performanceLevel(a)
	return a
}

func performanceLevel(a Analysis) string {
	switch {
	case a.Attempted == 0:
		return "INCOMPLETE"
	case a.Accuracy >= 75:
		return "STRONG"
	case a.Accuracy >= 50:
		return "AVERAGE"
	default:
		return "NEEDS_IMPROVEMENT"
	}
}

// Recommendations are rule-based (weak topic -> fixed study advice). This is not AI.
var topicAdvice = map[string]string{
	"Data Structures":   "Practice stacks, queues, trees and graphs, and be able to state the time complexity of common algorithms.",
	"Java":              "Revisit collections, String handling and multithreading basics such as volatile and synchronized.",
	"OOP":               "Review the four pillars of OOP and the SOLID principles with small code examples.",
	"DBMS":              "Review normalization, keys, ACID properties and transaction isolation levels.",
	"SQL":               "Practice joins, GROUP BY with HAVING, subqueries and indexing on sample tables.",
	"Operating Systems": "Focus on process scheduling, synchronization and deadlock concepts.",
	"Computer Networks": "Review the TCP/IP model, TCP vs UDP, the handshake and common application protocols.",
	"System Design":     "Study load balancing, caching, the CAP theorem and consistent hashing.",
	"Programming":       "Practice recursion and core problem-solving patterns.",
}

func Recommendations(a Analysis) []string {
	out := []string{}
	if a.Attempted == 0 {
		return append(out, "No questions were answered. Retake the assessment to get a full report.")
	}
	for _, topic := range a.Weak {
		if advice, ok := topicAdvice[topic]; ok {
			out = append(out, topic+": "+advice)
		} else {
			out = append(out, topic+": review the fundamentals and practise more questions.")
		}
	}
	if len(out) == 0 {
		out = append(out, "No weak topics found. Try assessments with more questions to be tested at harder levels.")
	}
	return out
}

func pct(correct, total int) float64 {
	if total == 0 {
		return 0
	}
	return round1(float64(correct) * 100 / float64(total))
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

func totalPages(total int64, size int) int {
	return int(math.Ceil(float64(total) / float64(size)))
}