package service

import (
	"errors"
	"strings"
	"time"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Claims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	db        *gorm.DB
	secret    []byte
	expiryHrs int
}

func NewAuthService(db *gorm.DB, secret string, expiryHrs int) *AuthService {
	return &AuthService{db: db, secret: []byte(secret), expiryHrs: expiryHrs}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.UserDTO, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	var count int64
	s.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return nil, Conflict("email already registered")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// Public registration can ONLY create candidates; admins come from seed data.
	user := models.User{Name: req.Name, Email: email, PasswordHash: string(hash), Role: models.RoleCandidate}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}
	return toUserDTO(user), nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	var user models.User
	err := s.db.Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		(err == nil && bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil) {
		return nil, Unauthorized("invalid email or password") // same message for both cases
	}
	if err != nil {
		return nil, err
	}
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{Token: token, User: *toUserDTO(user)}, nil
}

func (s *AuthService) GetUser(id uint) (*dto.UserDTO, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, NotFound("user not found")
	}
	return toUserDTO(user), nil
}

func (s *AuthService) generateToken(u models.User) (string, error) {
	claims := Claims{
		UserID: u.ID,
		Role:   string(u.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.expiryHrs) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *AuthService) ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return nil, Unauthorized("invalid or expired token")
	}
	return claims, nil
}

func toUserDTO(u models.User) *dto.UserDTO {
	return &dto.UserDTO{ID: u.ID, Name: u.Name, Email: u.Email, Role: string(u.Role)}
}