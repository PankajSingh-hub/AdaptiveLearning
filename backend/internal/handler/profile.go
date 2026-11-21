package handler

import (
	"net/http"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct{ svc *service.ProfileService }

func NewProfileHandler(s *service.ProfileService) *ProfileHandler { return &ProfileHandler{svc: s} }

func (h *ProfileHandler) Get(c *gin.Context) {
	res, err := h.svc.Get(c.GetUint("userID"), c.Query("tz"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *ProfileHandler) Streak(c *gin.Context) {
	res, err := h.svc.Streak(c.GetUint("userID"), c.Query("tz"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *ProfileHandler) Update(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "name must be 2 to 100 characters")
		return
	}
	res, err := h.svc.UpdateName(c.GetUint("userID"), req.Name)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "profile updated", res)
}

func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "current password, a new password of 8+ characters and its confirmation are required")
		return
	}
	if err := h.svc.ChangePassword(c.GetUint("userID"), req); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "password updated", nil)
}