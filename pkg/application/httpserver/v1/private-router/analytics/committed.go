package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/utils"
)

// Committed returns the committed-money / free-money summary.
func (h *Handler) Committed(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	out, err := h.app.AnalyticsService().Committed(c.Request.Context(), uid)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	utils.WriteJSON(c.Writer, http.StatusOK, out)
}
