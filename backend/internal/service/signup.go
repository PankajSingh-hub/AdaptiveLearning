package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
	"unicode"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/mailer"
	"adaptive-assessment/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	codeTTL         = 10 * time.Minute
	signupWindow    = 15 * time.Minute // time to set a password after the code is verified
	resendCooldown  = 60 * time.Second
	maxSendsPerHour = 5
	maxCodeAttempts = 5
)

type SignupService struct {
	db        *gorm.DB
	mailer    mailer.Mailer
	signupKey []byte // separate from the login key, so the two token kinds are not interchangeable
	auth      *AuthService
}

func NewSignupService(db *gorm.DB, m mailer.Mailer, secret string, auth *AuthService) *SignupService {
	return &SignupService{db: db, mailer: m, signupKey: []byte(secret + "|signup-v1"), auth: auth}
}

func normalizeEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }

// Start creates a fresh code for the email and sends it.
func (s *SignupService) Start(rawEmail string) (*dto.SignupStartResponse, error) {
	email := normalizeEmail(rawEmail)
	var existing int64
	if err := s.db.Model(&models.User{}).Where("email = ?", email).Count(&existing).Error; err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, Conflict("an account with this email already exists, please sign in")
	}

	now := time.Now()
	var v models.EmailVerification
	found := s.db.Where("email = ?", email).First(&v).Error == nil
	if found && now.Sub(v.CreatedAt) < time.Hour {
		if wait := resendCooldown - now.Sub(v.LastSentAt); wait > 0 {
			return nil, TooMany(fmt.Sprintf("please wait %d seconds before requesting another code", int(wait.Seconds())+1))
		}
		if v.SendCount >= maxSendsPerHour {
			return nil, TooMany("too many codes requested for this email, try again in an hour")
		}
		v.SendCount++
	} else {
		v.SendCount, v.CreatedAt = 1, now // a new one-hour window
	}

	code, err := newCode()
	if err != nil {
		return nil, err
	}
	v.Email, v.CodeHash = email, s.hashCode(email, code)
	v.ExpiresAt, v.LastSentAt, v.Attempts, v.VerifiedAt = now.Add(codeTTL), now, 0, nil
	if found {
		err = s.db.Save(&v).Error
	} else {
		err = s.db.Create(&v).Error
	}
	if err != nil {
		return nil, TooMany("a code was just requested for this email, please try again in a moment")
	}

	body := fmt.Sprintf("Your AdaptIQ verification code is: %s\n\nIt expires in %d minutes. If you did not request it, you can ignore this email.\n",
		code, int(codeTTL.Minutes()))
	if err := s.mailer.Send(email, "Your AdaptIQ verification code", body); err != nil {
		log.Printf("signup: sending the verification email failed: %v", err)
		s.db.Model(&v).Update("last_sent_at", now.Add(-resendCooldown)) // let the user retry right away
		return nil, BadGateway("could not send the verification email, please try again shortly")
	}
	log.Println("signup: verification code sent")
	return &dto.SignupStartResponse{ResendAfterSec: int(resendCooldown.Seconds()), ExpiresInMin: int(codeTTL.Minutes())}, nil
}

// Verify checks the code and returns a short-lived signup token.
func (s *SignupService) Verify(rawEmail, code string) (*dto.SignupVerifyResponse, error) {
	email := normalizeEmail(rawEmail)
	var verdict error // returned AFTER commit, so a wrong-attempt counter is saved instead of rolled back
	var token string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var v models.EmailVerification
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("email = ?", email).First(&v).Error; err != nil {
			verdict = BadRequest("no code was requested for this email, please request one")
			return nil
		}
		now := time.Now()
		switch {
		case now.After(v.ExpiresAt):
			verdict = BadRequest("this code has expired, please request a new one")
			return nil
		case v.Attempts >= maxCodeAttempts:
			verdict = BadRequest("too many wrong attempts, please request a new code")
			return nil
		case !hmac.Equal([]byte(s.hashCode(email, code)), []byte(v.CodeHash)):
			v.Attempts++
			if left := maxCodeAttempts - v.Attempts; left > 0 {
				verdict = BadRequest(fmt.Sprintf("incorrect code, %d attempt(s) left", left))
			} else {
				verdict = BadRequest("incorrect code, please request a new one")
			}
			return tx.Model(&v).Update("attempts", v.Attempts).Error
		}
		v.VerifiedAt, v.ExpiresAt = &now, now.Add(signupWindow)
		if err := tx.Save(&v).Error; err != nil {
			return err
		}
		t, err := s.signupToken(email)
		token = t
		return err
	})
	if err != nil {
		return nil, err
	}
	if verdict != nil {
		return nil, verdict
	}
	return &dto.SignupVerifyResponse{SignupToken: token}, nil
}

// Complete creates the account once the email is verified, then signs the user in.
func (s *SignupService) Complete(req dto.SignupCompleteRequest) (*dto.LoginResponse, error) {
	email, err := s.parseSignupToken(req.SignupToken)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if len(name) < 2 {
		return nil, BadRequest("please enter your name")
	}
	if req.Password != req.ConfirmPassword {
		return nil, BadRequest("passwords do not match")
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var v models.EmailVerification
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("email = ?", email).First(&v).Error; err != nil {
			return BadRequest("verification expired, please start again")
		}
		now := time.Now()
		if v.VerifiedAt == nil || now.After(v.ExpiresAt) {
			return BadRequest("verification expired, please start again")
		}
		var n int64
		if err := tx.Model(&models.User{}).Where("email = ?", email).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return Conflict("an account with this email already exists, please sign in")
		}
		user = models.User{Name: name, Email: email, PasswordHash: string(hash), Role: models.RoleCandidate, EmailVerifiedAt: &now}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Delete(&v).Error // the verification is single-use
	})
	if err != nil {
		return nil, err
	}
	token, err := s.auth.generateToken(user)
	if err != nil {
		return nil, err
	}
	log.Printf("signup: account created user=%d", user.ID)
	return &dto.LoginResponse{Token: token, User: *toUserDTO(user)}, nil
}

// ---- helpers ----

type signupClaims struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

func (s *SignupService) signupToken(email string) (string, error) {
	c := signupClaims{Email: email, Purpose: "signup", RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(signupWindow)), IssuedAt: jwt.NewNumericDate(time.Now()),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.signupKey)
}

func (s *SignupService) parseSignupToken(raw string) (string, error) {
	c := &signupClaims{}
	tok, err := jwt.ParseWithClaims(raw, c, func(*jwt.Token) (interface{}, error) { return s.signupKey, nil },
		jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !tok.Valid || c.Purpose != "signup" || c.Email == "" {
		return "", BadRequest("verification expired, please start again") // 400, not 401: a 401 would trigger the global logout redirect
	}
	return c.Email, nil
}

func (s *SignupService) hashCode(email, code string) string {
	m := hmac.New(sha256.New, s.signupKey)
	m.Write([]byte(email + "|" + code))
	return hex.EncodeToString(m.Sum(nil))
}

func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func validatePassword(pw string) error {
	if len(pw) < 8 || len(pw) > 72 { // 72 bytes is bcrypt's limit
		return BadRequest("password must be 8 to 72 characters")
	}
	var letter, digit bool
	for _, r := range pw {
		letter = letter || unicode.IsLetter(r)
		digit = digit || unicode.IsDigit(r)
	}
	if !letter || !digit {
		return BadRequest("password must contain at least one letter and one number")
	}
	return nil
}