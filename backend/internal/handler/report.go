package handler

import (
	"net/http"
	"strconv"

	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct{ svc *service.ReportService }

func NewReportHandler(s *service.ReportService) *ReportHandler { return &ReportHandler{svc: s} }

func (h *ReportHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	candidateID, _ := strconv.ParseUint(c.Query("candidateId"), 10, 64) // honoured for admins only
	res, err := h.svc.List(c.GetUint("userID"), isAdmin(c), uint(candidateID), page, size)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *ReportHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res, err := h.svc.GetByID(id, c.GetUint("userID"), isAdmin(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}