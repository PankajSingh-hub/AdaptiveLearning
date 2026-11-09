package service

import (
	"math"
	"strings"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var validCategories = map[string]bool{
	"DSA": true, "Java": true, "OOP": true, "DBMS": true, "OS": true,
	"Computer Networks": true, "System Design": true, "SQL": true, "Programming": true,
}

type QuestionFilter struct {
	Search, Topic, Difficulty, Status string // Status: "active", "inactive" or ""
	Page, Size                        int
}

type QuestionService struct{ db *gorm.DB }

func NewQuestionService(db *gorm.DB) *QuestionService { return &QuestionService{db: db} }

func orderedOptions(d *gorm.DB) *gorm.DB { return d.Order("question_options.id") }

func (s *QuestionService) List(f QuestionFilter) (*dto.PagedQuestions, error) {
	if f.Page < 0 {
		f.Page = 0
	}
	if f.Size < 1 || f.Size > 100 {
		f.Size = 10
	}
	q := s.db.Model(&models.Question{}).Joins("JOIN topics ON topics.id = questions.topic_id")
	if f.Search != "" {
		q = q.Where("questions.question_text LIKE ?", "%"+f.Search+"%")
	}
	if f.Topic != "" {
		q = q.Where("topics.name = ?", f.Topic)
	}
	if f.Difficulty != "" {
		q = q.Where("questions.difficulty = ?", strings.ToUpper(f.Difficulty))
	}
	switch f.Status {
	case "active":
		q = q.Where("questions.active = ?", true)
	case "inactive":
		q = q.Where("questions.active = ?", false)
	}
	base := q.Session(&gorm.Session{})

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []models.Question
	err := base.Select("questions.*").
		Preload("Options", orderedOptions).Preload("Topic").
		Order("questions.id DESC").Offset(f.Page * f.Size).Limit(f.Size).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]dto.AdminQuestionDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, toAdminDTO(r))
	}
	return &dto.PagedQuestions{
		Items: items, Page: f.Page, Size: f.Size, TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(f.Size))),
	}, nil
}

func (s *QuestionService) Get(id uint) (*dto.AdminQuestionDTO, error) {
	var q models.Question
	if err := s.db.Preload("Options", orderedOptions).Preload("Topic").First(&q, id).Error; err != nil {
		return nil, NotFound("question not found")
	}
	d := toAdminDTO(q)
	return &d, nil
}

func (s *QuestionService) Create(r dto.QuestionRequest) (*dto.AdminQuestionDTO, error) {
	if err := validateQuestion(r); err != nil {
		return nil, err
	}
	var result models.Question
	err := s.db.Transaction(func(tx *gorm.DB) error {
		topic, err := findOrCreateTopic(tx, r.Topic)
		if err != nil {
			return err
		}
		q := models.Question{
			QuestionText: strings.TrimSpace(r.Text), TopicID: topic.ID, Category: r.Category,
			Difficulty: models.Difficulty(r.Difficulty), QuestionType: "MCQ",
			Explanation: r.Explanation, TimeLimitSec: timeLimit(r.TimeLimitSec),
			Active: r.Active == nil || *r.Active,
		}
		if err := tx.Create(&q).Error; err != nil {
			return err
		}
		opts := buildOptions(q.ID, r.Options)
		if err := tx.Create(&opts).Error; err != nil {
			return err
		}
		q.Options, q.Topic = opts, *topic
		result = q
		return nil
	})
	if err != nil {
		return nil, err
	}
	d := toAdminDTO(result)
	return &d, nil
}

func (s *QuestionService) Update(id uint, r dto.QuestionRequest) (*dto.AdminQuestionDTO, error) {
	if err := validateQuestion(r); err != nil {
		return nil, err
	}
	var q models.Question
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Options", orderedOptions).First(&q, id).Error; err != nil {
			return NotFound("question not found")
		}
		topic, err := findOrCreateTopic(tx, r.Topic)
		if err != nil {
			return err
		}
		var used int64
		tx.Model(&models.SessionQuestion{}).Where("question_id = ?", id).Count(&used)
		if used > 0 && !sameOptions(q.Options, r.Options) {
			return BadRequest("this question was already used in interviews, so its options cannot change; deactivate it and create a new one")
		}
		q.QuestionText, q.TopicID, q.Category = strings.TrimSpace(r.Text), topic.ID, r.Category
		q.Difficulty, q.Explanation, q.TimeLimitSec = models.Difficulty(r.Difficulty), r.Explanation, timeLimit(r.TimeLimitSec)
		if r.Active != nil {
			q.Active = *r.Active
		}
		if err := tx.Omit(clause.Associations).Save(&q).Error; err != nil {
			return err
		}
		if used == 0 {
			if err := tx.Where("question_id = ?", id).Delete(&models.QuestionOption{}).Error; err != nil {
				return err
			}
			opts := buildOptions(id, r.Options)
			if err := tx.Create(&opts).Error; err != nil {
				return err
			}
			q.Options = opts
		}
		q.Topic = *topic
		return nil
	})
	if err != nil {
		return nil, err
	}
	d := toAdminDTO(q)
	return &d, nil
}

// Delete removes a question. If it was ever served in an interview it is only
// deactivated, so historical responses and reports stay intact.
func (s *QuestionService) Delete(id uint) (deactivated bool, err error) {
	var q models.Question
	if err := s.db.First(&q, id).Error; err != nil {
		return false, NotFound("question not found")
	}
	var used int64
	s.db.Model(&models.SessionQuestion{}).Where("question_id = ?", id).Count(&used)
	if used > 0 {
		return true, s.db.Model(&q).Update("active", false).Error
	}
	return false, s.db.Delete(&q).Error // options are removed by ON DELETE CASCADE
}

// ---- helpers ----

func validateQuestion(r dto.QuestionRequest) error {
	if !validCategories[r.Category] {
		return BadRequest("invalid category")
	}
	correct := 0
	for _, o := range r.Options {
		if strings.TrimSpace(o.Text) == "" {
			return BadRequest("option text cannot be blank")
		}
		if o.Correct {
			correct++
		}
	}
	if correct != 1 {
		return BadRequest("exactly one option must be marked correct")
	}
	return nil
}

func findOrCreateTopic(tx *gorm.DB, name string) (*models.Topic, error) {
	var t models.Topic
	err := tx.Where(models.Topic{Name: strings.TrimSpace(name)}).FirstOrCreate(&t).Error
	return &t, err
}

func buildOptions(questionID uint, in []dto.OptionInput) []models.QuestionOption {
	opts := make([]models.QuestionOption, 0, len(in))
	for _, o := range in {
		opts = append(opts, models.QuestionOption{QuestionID: questionID, OptionText: strings.TrimSpace(o.Text), IsCorrect: o.Correct})
	}
	return opts
}

func sameOptions(existing []models.QuestionOption, in []dto.OptionInput) bool {
	if len(existing) != len(in) {
		return false
	}
	for i := range existing {
		if existing[i].OptionText != strings.TrimSpace(in[i].Text) || existing[i].IsCorrect != in[i].Correct {
			return false
		}
	}
	return true
}

func timeLimit(v int) int {
	if v == 0 {
		return 60
	}
	return v
}

func toAdminDTO(q models.Question) dto.AdminQuestionDTO {
	opts := make([]dto.AdminOptionDTO, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, dto.AdminOptionDTO{ID: o.ID, Text: o.OptionText, Correct: o.IsCorrect})
	}
	return dto.AdminQuestionDTO{
		ID: q.ID, Text: q.QuestionText, Topic: q.Topic.Name, Category: q.Category,
		Difficulty: string(q.Difficulty), QuestionType: q.QuestionType, Explanation: q.Explanation,
		TimeLimitSec: q.TimeLimitSec, Active: q.Active, Options: opts, CreatedAt: q.CreatedAt,
	}
}