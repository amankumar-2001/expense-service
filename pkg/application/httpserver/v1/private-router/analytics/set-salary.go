package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/pkg/domain/dto/request"
	"github.com/kharchibook/expense-service/pkg/domain/dto/response"
	"github.com/kharchibook/expense-service/utils"
)

// SetSalary stores the user's monthly salary and salary day.
func (h *Handler) SetSalary(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	var req request.SetSalaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WriteError(c.Writer, apperrors.BadRequestError("invalid request body"))
		return
	}
	if err := req.Validate(); err != nil {
		utils.WriteError(c.Writer, apperrors.ValidationError(err))
		return
	}
	salary, err := h.app.FinanceService().Set(c.Request.Context(), uid, req.Amount, req.SalaryDay)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	utils.WriteJSON(c.Writer, http.StatusOK, response.SalaryResponse{
		MonthlySalary: salary.MonthlySalary,
		SalaryDay:     salary.SalaryDay,
	})
}
