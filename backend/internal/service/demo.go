package service

import (
	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"
)

const DemoEmail = "candidate@example.com"

// DemoLogin signs in as the shared demo candidate. It never returns an admin.
func (s *AuthService) DemoLogin() (*dto.LoginResponse, error) {
	var u models.User
	if err := s.db.Where("email = ? AND role = ?", DemoEmail, models.RoleCandidate).First(&u).Error; err != nil {
		return nil, NotFound("the demo account is not available")
	}
	token, err := s.generateToken(u)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{Token: token, User: *toUserDTO(u)}, nil
}