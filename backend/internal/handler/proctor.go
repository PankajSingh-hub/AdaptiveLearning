package handler

import (
	"net/http"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *InterviewHandler) Warn(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.WarningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "type must be NO_PERSON or MULTIPLE_PEOPLE")
		return
	}
	res, err := h.svc.Warn(c.GetUint("userID"), id, req.Type)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warning recorded", res)
}

func (h *CodingHandler) Warn(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.WarningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "type must be NO_PERSON or MULTIPLE_PEOPLE")
		return
	}
	res, err := h.svc.Warn(c.GetUint("userID"), id, req.Type)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warning recorded", res)
}