package handler

import (
	"net/http"
	"strconv"

	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type CandidateHandler struct{ svc *service.CandidateService }

func NewCandidateHandler(s *service.CandidateService) *CandidateHandler {
	return &CandidateHandler{svc: s}
}

func (h *CandidateHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	res, err := h.svc.List(c.Query("search"), c.Query("sort"), c.Query("dir"), page, size)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *CandidateHandler) Get(c *gin.Context) {
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