package expense

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/utils"
)

// List returns the user's expenses for a month (query: month=YYYY-MM, category).
func (h *Handler) List(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	out, err := h.app.ExpenseService().List(c.Request.Context(), uid, c.Query("month"), c.Query("category"))
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	utils.WriteJSON(c.Writer, http.StatusOK, out)
}
