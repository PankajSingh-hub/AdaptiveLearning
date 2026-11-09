package service

import (
	"testing"

	"adaptive-assessment/internal/dto"
)

func validRequest() dto.QuestionRequest {
	return dto.QuestionRequest{
		Text: "What is the time complexity of binary search?", Topic: "Data Structures",
		Category: "DSA", Difficulty: "MEDIUM",
		Options: []dto.OptionInput{{Text: "O(n)"}, {Text: "O(log n)", Correct: true}},
	}
}

func TestValidQuestionPasses(t *testing.T) {
	if err := validateQuestion(validRequest()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQuestionNeedsExactlyOneCorrectOption(t *testing.T) {
	none := validRequest()
	none.Options[1].Correct = false
	if validateQuestion(none) == nil {
		t.Fatal("zero correct options must fail")
	}
	two := validRequest()
	two.Options[0].Correct = true
	if validateQuestion(two) == nil {
		t.Fatal("two correct options must fail")
	}
}

func TestQuestionRejectsUnknownCategory(t *testing.T) {
	r := validRequest()
	r.Category = "Astrology"
	if validateQuestion(r) == nil {
		t.Fatal("unknown category must fail")
	}
}

func TestQuestionRejectsBlankOption(t *testing.T) {
	r := validRequest()
	r.Options[0].Text = "   "
	if validateQuestion(r) == nil {
		t.Fatal("blank option text must fail")
	}
}