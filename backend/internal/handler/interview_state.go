package handler

import (
	"net/http"

	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *InterviewHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res, err := h.svc.Get(c.GetUint("userID"), id)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *InterviewHandler) Finish(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res, err := h.svc.Finish(c.GetUint("userID"), id)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "interview finished", res)
}