package expense

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/pkg/domain/dto/request"
	"github.com/kharchibook/expense-service/utils"
)

// Create logs a new expense.
func (h *Handler) Create(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	var req request.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WriteError(c.Writer, apperrors.BadRequestError("invalid request body"))
		return
	}
	if err := req.Validate(); err != nil {
		utils.WriteError(c.Writer, apperrors.ValidationError(err))
		return
	}
	out, err := h.app.ExpenseService().Create(c.Request.Context(), uid, req)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	utils.WriteJSON(c.Writer, http.StatusCreated, out)
}
