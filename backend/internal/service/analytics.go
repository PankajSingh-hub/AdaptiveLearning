package service

import (
	"sort"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
)

type AnalyticsService struct {
	db      *gorm.DB
	reports *ReportService
}

func NewAnalyticsService(db *gorm.DB, r *ReportService) *AnalyticsService {
	return &AnalyticsService{db: db, reports: r}
}

type statRow struct {
	Name      string
	Attempted int
	Correct   int
}

const candidateResponsesFrom = ` FROM responses r
	JOIN session_questions sq ON sq.id = r.session_question_id
	JOIN interview_sessions s ON s.id = sq.session_id
	JOIN questions q ON q.id = sq.question_id
	JOIN topics t ON t.id = q.topic_id
	WHERE s.candidate_id = ?`

// Candidate computes everything from the responses table with SQL aggregates.
func (s *AnalyticsService) Candidate(candidateID uint) (*dto.CandidateAnalytics, error) {
	var user models.User
	if err := s.db.Where("id = ? AND role = ?", candidateID, string(models.RoleCandidate)).First(&user).Error; err != nil {
		return nil, NotFound("candidate not found")
	}

	var tot struct {
		Attempted int
		Correct   int
		AvgTime   float64
		Score     int
	}
	err := s.db.Raw("SELECT COUNT(*) AS attempted, COALESCE(CAST(SUM(r.is_correct) AS SIGNED),0) AS correct, "+
		"COALESCE(AVG(r.time_taken_sec),0) AS avg_time, COALESCE(CAST(SUM(r.points_awarded) AS SIGNED),0) AS score"+
		candidateResponsesFrom, candidateID).Scan(&tot).Error
	if err != nil {
		return nil, err
	}

	var topicRows, diffRows []statRow
	err = s.db.Raw("SELECT t.name AS name, COUNT(*) AS attempted, CAST(SUM(r.is_correct) AS SIGNED) AS correct"+
		candidateResponsesFrom+" GROUP BY t.name", candidateID).Scan(&topicRows).Error
	if err != nil {
		return nil, err
	}
	err = s.db.Raw("SELECT sq.difficulty AS name, COUNT(*) AS attempted, CAST(SUM(r.is_correct) AS SIGNED) AS correct"+
		candidateResponsesFrom+" GROUP BY sq.difficulty", candidateID).Scan(&diffRows).Error
	if err != nil {
		return nil, err
	}

	trend := []dto.TrendPoint{}
	err = s.db.Raw("SELECT created_at AS taken_at, accuracy, score FROM interview_reports WHERE candidate_id = ? ORDER BY created_at",
		candidateID).Scan(&trend).Error
	if err != nil {
		return nil, err
	}
	recent, err := s.reports.List(candidateID, false, 0, 0, 5)
	if err != nil {
		return nil, err
	}

	topics := topicStats(topicRows) // sorted best -> worst
	scoreSum := 0
	for _, p := range trend {
		scoreSum += p.Score
	}
	avgScore := 0.0
	if len(trend) > 0 {
		avgScore = round1(float64(scoreSum) / float64(len(trend)))
	}
	return &dto.CandidateAnalytics{
		CandidateID: user.ID, CandidateName: user.Name, CompletedInterviews: len(trend),
		QuestionsAttempted: tot.Attempted, CorrectAnswers: tot.Correct,
		IncorrectAnswers: tot.Attempted - tot.Correct, Accuracy: pct(tot.Correct, tot.Attempted),
		TotalScore: tot.Score, AvgScore: avgScore, AvgResponseSec: round1(tot.AvgTime),
		TopicPerformance: topics, DifficultyPerformance: difficultyStats(diffRows),
		Trend: trend, RecommendedTopics: weakestTopics(topics, 3), RecentReports: recent.Items,
	}, nil
}

func (s *AnalyticsService) Admin() (*dto.AdminAnalytics, error) {
	out := &dto.AdminAnalytics{}
	s.db.Model(&models.User{}).Where("role = ?", string(models.RoleCandidate)).Count(&out.TotalCandidates)
	s.db.Model(&models.Assessment{}).Count(&out.TotalAssessments)
	s.db.Model(&models.Question{}).Count(&out.TotalQuestions)
	s.db.Model(&models.InterviewReport{}).Count(&out.CompletedInterviews)

	var avg struct{ AvgScore, AvgAccuracy float64 }
	if err := s.db.Raw("SELECT COALESCE(AVG(score),0) AS avg_score, COALESCE(AVG(accuracy),0) AS avg_accuracy FROM interview_reports").
		Scan(&avg).Error; err != nil {
		return nil, err
	}
	out.AvgScore, out.AvgAccuracy = round1(avg.AvgScore), round1(avg.AvgAccuracy)

	out.DifficultyDistribution = []dto.LabelCount{}
	if err := s.db.Raw("SELECT difficulty AS label, COUNT(*) AS total FROM questions WHERE active = 1 GROUP BY difficulty").
		Scan(&out.DifficultyDistribution).Error; err != nil {
		return nil, err
	}
	out.TopicDistribution = []dto.LabelCount{}
	if err := s.db.Raw("SELECT t.name AS label, COUNT(*) AS total FROM questions q JOIN topics t ON t.id = q.topic_id " +
		"WHERE q.active = 1 GROUP BY t.name ORDER BY total DESC").Scan(&out.TopicDistribution).Error; err != nil {
		return nil, err
	}
	out.TopCandidates = []dto.TopCandidate{}
	if err := s.db.Raw("SELECT u.id AS candidate_id, u.name AS name, COUNT(r.id) AS interviews, " +
		"AVG(r.accuracy) AS avg_accuracy, AVG(r.score) AS avg_score " +
		"FROM interview_reports r JOIN users u ON u.id = r.candidate_id " +
		"GROUP BY u.id, u.name ORDER BY avg_accuracy DESC, avg_score DESC LIMIT 5").Scan(&out.TopCandidates).Error; err != nil {
		return nil, err
	}
	for i := range out.TopCandidates {
		out.TopCandidates[i].AvgAccuracy = round1(out.TopCandidates[i].AvgAccuracy)
		out.TopCandidates[i].AvgScore = round1(out.TopCandidates[i].AvgScore)
	}
	recent, err := s.reports.List(0, true, 0, 0, 5)
	if err != nil {
		return nil, err
	}
	out.RecentInterviews = recent.Items
	return out, nil
}

func topicStats(rows []statRow) []dto.TopicStatDTO {
	out := make([]dto.TopicStatDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.TopicStatDTO{Topic: r.Name, Attempted: r.Attempted, Correct: r.Correct, Accuracy: pct(r.Correct, r.Attempted)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Accuracy != out[j].Accuracy {
			return out[i].Accuracy > out[j].Accuracy
		}
		return out[i].Topic < out[j].Topic
	})
	return out
}

func difficultyStats(rows []statRow) []dto.DifficultyStatDTO {
	byName := map[string]statRow{}
	for _, r := range rows {
		byName[r.Name] = r
	}
	out := []dto.DifficultyStatDTO{}
	for _, d := range []models.Difficulty{models.Easy, models.Medium, models.Hard} {
		if r, ok := byName[string(d)]; ok {
			out = append(out, dto.DifficultyStatDTO{Difficulty: string(d), Attempted: r.Attempted, Correct: r.Correct, Accuracy: pct(r.Correct, r.Attempted)})
		}
	}
	return out
}

// weakestTopics expects topics sorted best -> worst and returns up to n topics below 75%.
func weakestTopics(topics []dto.TopicStatDTO, n int) []string {
	out := []string{}
	for i := len(topics) - 1; i >= 0 && len(out) < n; i-- {
		if topics[i].Accuracy < strongThreshold {
			out = append(out, topics[i].Topic)
		}
	}
	return out
}