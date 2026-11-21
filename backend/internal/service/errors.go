package service

import (
	"errors"
	"net/http"

	"github.com/go-sql-driver/mysql"
)

// AppError carries an HTTP status so handlers never leak raw errors.
type AppError struct {
	Status  int
	Message string
}

func (e *AppError) Error() string { return e.Message }

func NotFound(msg string) *AppError   { return &AppError{http.StatusNotFound, msg} }
func BadRequest(msg string) *AppError { return &AppError{http.StatusBadRequest, msg} }
func Unauthorized(msg string) *AppError {
	return &AppError{http.StatusUnauthorized, msg}
}
func Conflict(msg string) *AppError { return &AppError{http.StatusConflict, msg} }
func TooMany(msg string) *AppError    { return &AppError{http.StatusTooManyRequests, msg} }
func BadGateway(msg string) *AppError { return &AppError{http.StatusBadGateway, msg} }
func Forbidden(msg string) *AppError { return &AppError{http.StatusForbidden, msg} }
// isDuplicateKey reports a violated UNIQUE constraint (MySQL error 1062).
func isDuplicateKey(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}