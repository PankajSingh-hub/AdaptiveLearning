package service

import (
	"strings"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
)

type CandidateService struct {
	db      *gorm.DB
	reports *ReportService
}

func NewCandidateService(db *gorm.DB, r *ReportService) *CandidateService {
	return &CandidateService{db: db, reports: r}
}

// Whitelist: the sort column is NEVER taken directly from user input (prevents SQL injection).
var candidateSorts = map[string]string{
	"name": "u.name", "email": "u.email", "joined": "u.created_at",
	"interviews": "interviews", "accuracy": "avg_accuracy",
}

func (s *CandidateService) List(search, sortBy, dir string, page, size int) (*dto.PagedCandidates, error) {
	if page < 0 {
		page = 0
	}
	if size < 1 || size > 100 {
		size = 10
	}
	sortCol, ok := candidateSorts[sortBy]
	if !ok {
		sortCol, dir = "u.created_at", "desc"
	}
	direction := "DESC"
	if strings.EqualFold(dir, "asc") {
		direction = "ASC"
	}

	where := "u.role = ?"
	args := []interface{}{string(models.RoleCandidate)}
	if search != "" {
		where += " AND (u.name LIKE ? OR u.email LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	var total int64
	if err := s.db.Table("users u").Where(where, args...).Count(&total).Error; err != nil {
		return nil, err
	}
	query := "SELECT u.id, u.name, u.email, u.created_at AS joined_at, COUNT(r.id) AS interviews, " +
		"COALESCE(AVG(r.accuracy),0) AS avg_accuracy, MAX(r.created_at) AS last_interview " +
		"FROM users u LEFT JOIN interview_reports r ON r.candidate_id = u.id WHERE " + where +
		" GROUP BY u.id, u.name, u.email, u.created_at ORDER BY " + sortCol + " " + direction + ", u.id LIMIT ? OFFSET ?"
	items := []dto.CandidateListItem{}
	queryArgs := append(append([]interface{}{}, args...), size, page*size)
	if err := s.db.Raw(query, queryArgs...).Scan(&items).Error; err != nil {
		return nil, err
	}
	for i := range items {
		items[i].AvgAccuracy = round1(items[i].AvgAccuracy)
	}
	return &dto.PagedCandidates{Items: items, Page: page, Size: size, TotalItems: total, TotalPages: totalPages(total, size)}, nil
}

func (s *CandidateService) Get(id uint) (*dto.CandidateDetail, error) {
	var u models.User
	if err := s.db.Where("id = ? AND role = ?", id, string(models.RoleCandidate)).First(&u).Error; err != nil {
		return nil, NotFound("candidate not found")
	}
	reports, err := s.reports.List(0, true, id, 0, 50)
	if err != nil {
		return nil, err
	}
	return &dto.CandidateDetail{ID: u.ID, Name: u.Name, Email: u.Email, JoinedAt: u.CreatedAt, Reports: reports.Items}, nil
}