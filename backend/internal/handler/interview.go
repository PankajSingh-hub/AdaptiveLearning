package handler

import (
	"net/http"
	"strconv"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type InterviewHandler struct{ svc *service.InterviewService }

func NewInterviewHandler(s *service.InterviewService) *InterviewHandler {
	return &InterviewHandler{svc: s}
}

func (h *InterviewHandler) Start(c *gin.Context) {
	var req dto.StartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "assessmentId is required")
		return
	}
	res, err := h.svc.Start(c.GetUint("userID"), req.AssessmentID)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "interview started", res)
}

func (h *InterviewHandler) Answer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid session id")
		return
	}
	var req dto.AnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "sessionQuestionId and optionId are required")
		return
	}
	res, err := h.svc.Answer(c.GetUint("userID"), uint(id), req)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "answer submitted successfully", res)
}