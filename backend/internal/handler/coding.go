package handler

import (
	"net/http"
	"strconv"

	"adaptive-assessment/internal/dto"
	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type CodingHandler struct{ svc *service.CodingService }

func NewCodingHandler(s *service.CodingService) *CodingHandler { return &CodingHandler{svc: s} }

func parseOrder(c *gin.Context) (int, bool) {
	n, err := strconv.Atoi(c.Param("order"))
	if err != nil || n < 1 {
		utils.Fail(c, http.StatusBadRequest, "invalid problem number")
		return 0, false
	}
	return n, true
}

func (h *CodingHandler) Languages(c *gin.Context) {
	utils.Success(c, http.StatusOK, "ok", h.svc.Languages())
}

func (h *CodingHandler) Start(c *gin.Context) {
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
	utils.Success(c, http.StatusCreated, "coding session ready", res)
}

func (h *CodingHandler) Session(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res, err := h.svc.GetSession(c.GetUint("userID"), id)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *CodingHandler) Problem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	order, ok := parseOrder(c)
	if !ok {
		return
	}
	res, err := h.svc.GetProblem(c.GetUint("userID"), id, order)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ok", res)
}

func (h *CodingHandler) Run(c *gin.Context)    { h.judge(c, false) }
func (h *CodingHandler) Submit(c *gin.Context) { h.judge(c, true) }

func (h *CodingHandler) judge(c *gin.Context, submit bool) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	order, ok := parseOrder(c)
	if !ok {
		return
	}
	var req dto.CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "language (cpp, java or python) and code are required")
		return
	}
	var res *dto.CodingResult
	var err error
	if submit {
		res, err = h.svc.Submit(c.GetUint("userID"), id, order, req)
	} else {
		res, err = h.svc.Run(c.GetUint("userID"), id, order, req)
	}
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "judged", res)
}

func (h *CodingHandler) Finish(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Finish(c.GetUint("userID"), id); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "coding session finished", nil)
}
func (h *CodingHandler) Skip(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	order, ok := parseOrder(c)
	if !ok {
		return
	}
	res, err := h.svc.Skip(c.GetUint("userID"), id, order)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "problem skipped", res)
}