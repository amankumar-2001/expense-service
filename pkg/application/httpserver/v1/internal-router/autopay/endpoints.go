package autopay

import (
	"github.com/gin-gonic/gin"
	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/domain/dto/request"
	"github.com/kharchibook/expense-service/utils"
)

// List returns the user's autopays (query: userId, status, type).
func (h *Handler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	out, err := h.app.AutoPayService().List(c.Request.Context(), uid, c.Query("status"), c.Query("type"))
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	writeOK(c, out)
}

// CreateDetected stores a mailbox-detected commitment as a pending entry awaiting
// confirmation (query: userId).
func (h *Handler) CreateDetected(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var req request.CreateDetectedAutoPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WriteError(c.Writer, apperrors.BadRequestError("invalid request body"))
		return
	}
	if err := req.Validate(); err != nil {
		utils.WriteError(c.Writer, apperrors.ValidationError(err))
		return
	}
	out, err := h.app.AutoPayService().CreateDetected(c.Request.Context(), uid, req)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	writeOK(c, out)
}

// Update applies a partial change to an autopay (query: userId).
func (h *Handler) Update(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req request.UpdateAutoPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WriteError(c.Writer, apperrors.BadRequestError("invalid request body"))
		return
	}
	if err := req.Validate(); err != nil {
		utils.WriteError(c.Writer, apperrors.ValidationError(err))
		return
	}
	out, err := h.app.AutoPayService().Update(c.Request.Context(), uid, id, req)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	writeOK(c, out)
}

// Confirm activates an auto-detected autopay (query: userId).
func (h *Handler) Confirm(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	out, err := h.app.AutoPayService().Confirm(c.Request.Context(), uid, id)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	writeOK(c, out)
}
