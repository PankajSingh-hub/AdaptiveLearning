package service

import (
	"log"
	"time"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const MaxWarnings = 3

// Warn records one proctoring warning for a quiz interview. The third warning ends the interview.
func (s *InterviewService) Warn(candidateID, sessionID uint, kind string) (*dto.WarningResponse, error) {
	res := &dto.WarningResponse{Max: MaxWarnings}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		sess, err := lockSession(tx, sessionID, candidateID)
		if err != nil {
			return err
		}
		res.Warnings, res.Terminated = sess.Warnings, sess.Terminated
		if sess.Status != models.InProgress {
			return nil
		}
		now := time.Now()
		if now.After(sess.ExpiresAt) {
			return closeSession(tx, sess, models.Expired, now)
		}
		sess.Warnings++
		res.Warnings = sess.Warnings
		if sess.Warnings >= MaxWarnings {
			sess.Terminated, res.Terminated = true, true
			return closeSession(tx, sess, models.Completed, now)
		}
		return tx.Save(sess).Error
	})
	if err != nil {
		return nil, err
	}
	log.Printf("proctor warning: interview session=%d type=%s count=%d", sessionID, kind, res.Warnings)
	if res.Terminated { // the report is generated after the transaction has committed
		if id, err := s.reports.Generate(sessionID); err != nil {
			log.Printf("report generation failed after termination: session=%d err=%v", sessionID, err)
		} else {
			res.ReportID = &id
		}
	}
	return res, nil
}

// Warn records one proctoring warning for a coding session. The third warning ends the session.
func (s *CodingService) Warn(candidateID, id uint, kind string) (*dto.WarningResponse, error) {
	if _, err := s.loadSession(candidateID, id); err != nil { // also closes an expired session
		return nil, err
	}
	res := &dto.WarningResponse{Max: MaxWarnings}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var sess models.CodingSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND candidate_id = ?", id, candidateID).First(&sess).Error; err != nil {
			return NotFound("coding session not found")
		}
		res.Warnings, res.Terminated = sess.Warnings, sess.Terminated
		if sess.Status != models.InProgress {
			return nil
		}
		sess.Warnings++
		res.Warnings = sess.Warnings
		over := sess.Warnings >= MaxWarnings
		if err := tx.Model(&models.CodingSession{}).Where("id = ?", id).
			Updates(map[string]any{"warnings": sess.Warnings, "is_terminated": over}).Error; err != nil {
			return err
		}
		if over {
			res.Terminated = true
			return closeCoding(tx, id, models.Completed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	log.Printf("proctor warning: coding session=%d type=%s count=%d", id, kind, res.Warnings)
	return res, nil
}