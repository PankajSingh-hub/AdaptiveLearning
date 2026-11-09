package handler

import (
	"net/http"

	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct{ svc *service.AnalyticsService }

func NewAnalyticsHandler(s *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: s}
}

// Candidate: admins can view anyone, candidates only themselves.
func (h *AnalyticsHandler) Candidate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if !isAdmin(c) && id != c.GetUint("userID") {
		utils.Fail(c, http.StatusForbidden, "you can only view your own analytics")
		return
	}
	res, err := h.svc.Candidate(id)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *AnalyticsHandler) Admin(c *gin.Context) {
	res, err := h.svc.Admin()
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}