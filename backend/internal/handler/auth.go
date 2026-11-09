package handler

import (
	"errors"
	"log"
	"net/http"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

// handleError converts service errors into clean API responses (no stack traces leak).
func handleError(c *gin.Context, err error) {
	var appErr *service.AppError
	if errors.As(err, &appErr) {
		utils.Fail(c, appErr.Status, appErr.Message)
		return
	}
	log.Printf("unexpected error: %v", err)
	utils.Fail(c, http.StatusInternalServerError, "something went wrong, please try again")
}

type AuthHandler struct{ auth *service.AuthService }

func NewAuthHandler(a *service.AuthService) *AuthHandler { return &AuthHandler{auth: a} }

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "validation failed: check name, email and password (min 8 chars)")
		return
	}
	user, err := h.auth.Register(req)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "registered successfully", user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "email and password are required")
		return
	}
	res, err := h.auth.Login(req)
	if err != nil {
		log.Printf("login failed for %s", req.Email)
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "login successful", res)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, err := h.auth.GetUser(c.GetUint("userID"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", user)
}