package expense

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/utils"
)

// Summary returns the monthly spend breakdown (query: month=YYYY-MM).
func (h *Handler) Summary(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	out, err := h.app.ExpenseService().Summary(c.Request.Context(), uid, c.Query("month"))
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	utils.WriteJSON(c.Writer, http.StatusOK, out)
}
