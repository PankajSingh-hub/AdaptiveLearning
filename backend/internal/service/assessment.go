package service

import (
	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
)

type AssessmentService struct{ db *gorm.DB }

func NewAssessmentService(db *gorm.DB) *AssessmentService { return &AssessmentService{db: db} }

// Candidates only see active assessments; admins see everything.
func (s *AssessmentService) List(adminView bool) ([]dto.AssessmentDTO, error) {
	q := s.db.Order("id DESC")
	if !adminView {
		q = q.Where("active = ?", true)
	}
	var rows []models.Assessment
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dto.AssessmentDTO, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAssessmentDTO(a))
	}
	return out, nil
}

func (s *AssessmentService) Get(id uint, adminView bool) (*dto.AssessmentDTO, error) {
	q := s.db
	if !adminView {
		q = q.Where("active = ?", true)
	}
	var a models.Assessment
	if err := q.First(&a, id).Error; err != nil {
		return nil, NotFound("assessment not found")
	}
	d := toAssessmentDTO(a)
	return &d, nil
}

func (s *AssessmentService) Create(r dto.AssessmentRequest) (*dto.AssessmentDTO, error) {
	a := models.Assessment{
		Title: r.Title, Description: r.Description, TotalQuestions: r.TotalQuestions,
		DurationMinutes: r.DurationMinutes, ShowExplanations: r.ShowExplanations,
		Active: r.Active == nil || *r.Active,
	}
	if err := s.db.Create(&a).Error; err != nil {
		return nil, err
	}
	d := toAssessmentDTO(a)
	return &d, nil
}

func (s *AssessmentService) Update(id uint, r dto.AssessmentRequest) (*dto.AssessmentDTO, error) {
	var a models.Assessment
	if err := s.db.First(&a, id).Error; err != nil {
		return nil, NotFound("assessment not found")
	}
	a.Title, a.Description, a.TotalQuestions = r.Title, r.Description, r.TotalQuestions
	a.DurationMinutes, a.ShowExplanations = r.DurationMinutes, r.ShowExplanations
	if r.Active != nil {
		a.Active = *r.Active
	}
	if err := s.db.Save(&a).Error; err != nil {
		return nil, err
	}
	d := toAssessmentDTO(a)
	return &d, nil
}

// Delete deactivates if candidates already took it, otherwise removes it.
func (s *AssessmentService) Delete(id uint) (deactivated bool, err error) {
	var a models.Assessment
	if err := s.db.First(&a, id).Error; err != nil {
		return false, NotFound("assessment not found")
	}
		var used, usedCoding int64
	s.db.Model(&models.InterviewSession{}).Where("assessment_id = ?", id).Count(&used)
	s.db.Model(&models.CodingSession{}).Where("assessment_id = ?", id).Count(&usedCoding)
	if used > 0 || usedCoding > 0 {
		return true, s.db.Model(&a).Update("active", false).Error
	}
	return false, s.db.Delete(&a).Error
}

func toAssessmentDTO(a models.Assessment) dto.AssessmentDTO {
	return dto.AssessmentDTO{
    ID: a.ID, Title: a.Title, Description: a.Description,
    Type: a.AssessmentType,
    TotalQuestions: a.TotalQuestions,
    DurationMinutes: a.DurationMinutes, ShowExplanations: a.ShowExplanations,
    Active: a.Active, CreatedAt: a.CreatedAt,
}
}