package handler

import (
	"net/http"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/models"
	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type AssessmentHandler struct{ svc *service.AssessmentService }

func NewAssessmentHandler(s *service.AssessmentService) *AssessmentHandler {
	return &AssessmentHandler{svc: s}
}

func isAdmin(c *gin.Context) bool { return c.GetString("role") == string(models.RoleAdmin) }

const assessmentValidationMsg = "validation failed: title, totalQuestions (1-50) and durationMinutes (1-240) are required"

func (h *AssessmentHandler) List(c *gin.Context) {
	res, err := h.svc.List(isAdmin(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *AssessmentHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res, err := h.svc.Get(id, isAdmin(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *AssessmentHandler) Create(c *gin.Context) {
	var req dto.AssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, assessmentValidationMsg)
		return
	}
	res, err := h.svc.Create(req)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "assessment created", res)
}

func (h *AssessmentHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.AssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, assessmentValidationMsg)
		return
	}
	res, err := h.svc.Update(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "assessment updated", res)
}

func (h *AssessmentHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	deactivated, err := h.svc.Delete(id)
	if err != nil {
		handleError(c, err)
		return
	}
	msg := "assessment deleted"
	if deactivated {
		msg = "assessment has interview history, so it was deactivated instead of deleted"
	}
	utils.Success(c, http.StatusOK, msg, nil)
}