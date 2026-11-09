package handler

import (
	"net/http"
	"strconv"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return uint(id), true
}

type QuestionHandler struct{ svc *service.QuestionService }

func NewQuestionHandler(s *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{svc: s}
}

const questionValidationMsg = "validation failed: text (min 10 chars), topic, category, difficulty (EASY/MEDIUM/HARD) and 2-6 options are required"

func (h *QuestionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	res, err := h.svc.List(service.QuestionFilter{
		Search: c.Query("search"), Topic: c.Query("topic"), Difficulty: c.Query("difficulty"),
		Status: c.Query("status"), Page: page, Size: size,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *QuestionHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res, err := h.svc.Get(id)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *QuestionHandler) Create(c *gin.Context) {
	var req dto.QuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, questionValidationMsg)
		return
	}
	res, err := h.svc.Create(req)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "question created", res)
}

func (h *QuestionHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.QuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, questionValidationMsg)
		return
	}
	res, err := h.svc.Update(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "question updated", res)
}

func (h *QuestionHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	deactivated, err := h.svc.Delete(id)
	if err != nil {
		handleError(c, err)
		return
	}
	msg := "question deleted"
	if deactivated {
		msg = "question was used in interviews, so it was deactivated instead of deleted"
	}
	utils.Success(c, http.StatusOK, msg, nil)
}