package models

import "time"

type EmailVerification struct {
	ID         uint `gorm:"primaryKey"`
	Email      string
	CodeHash   string
	ExpiresAt  time.Time
	Attempts   int
	SendCount  int
	LastSentAt time.Time
	VerifiedAt *time.Time
	CreatedAt  time.Time
}