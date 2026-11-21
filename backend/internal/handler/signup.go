package handler

import (
	"net/http"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type SignupHandler struct {
	svc         *service.SignupService
	auth        *service.AuthService
	demoEnabled bool
	emailMode   string
}

func NewSignupHandler(svc *service.SignupService, auth *service.AuthService, demoEnabled bool, emailMode string) *SignupHandler {
	return &SignupHandler{svc: svc, auth: auth, demoEnabled: demoEnabled, emailMode: emailMode}
}

func (h *SignupHandler) Config(c *gin.Context) {
	utils.Success(c, http.StatusOK, "ok", dto.AuthConfig{DemoLogin: h.demoEnabled, EmailMode: h.emailMode})
}

func (h *SignupHandler) Start(c *gin.Context) {
	var req dto.SignupStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "enter a valid email address")
		return
	}
	res, err := h.svc.Start(req.Email)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "verification code sent", res)
}

func (h *SignupHandler) Verify(c *gin.Context) {
	var req dto.SignupVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "enter the 6-digit code")
		return
	}
	res, err := h.svc.Verify(req.Email, req.Code)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "email verified", res)
}

func (h *SignupHandler) Complete(c *gin.Context) {
	var req dto.SignupCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "name, a password of 8+ characters and its confirmation are required")
		return
	}
	res, err := h.svc.Complete(req)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "account created", res)
}

func (h *SignupHandler) Demo(c *gin.Context) {
	if !h.demoEnabled {
		utils.Fail(c, http.StatusNotFound, "demo login is not enabled")
		return
	}
	res, err := h.auth.DemoLogin()
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "demo login successful", res)
}