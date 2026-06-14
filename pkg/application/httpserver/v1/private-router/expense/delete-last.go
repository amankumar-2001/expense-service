package expense

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/utils"
)

// DeleteLast removes the user's most recently logged expense.
func (h *Handler) DeleteLast(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	out, err := h.app.ExpenseService().DeleteLast(c.Request.Context(), uid)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	utils.WriteJSON(c.Writer, http.StatusOK, out)
}
