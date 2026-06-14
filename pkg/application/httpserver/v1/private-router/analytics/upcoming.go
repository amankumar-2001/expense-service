package analytics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/utils"
)

// Upcoming returns deductions due within the requested window (query: days,
// default 7).
func (h *Handler) Upcoming(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	days := defaultUpcomingDays
	if q := c.Query("days"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil || n < 0 || n > 366 {
			utils.WriteError(c.Writer, apperrors.BadRequestError("days must be between 0 and 366"))
			return
		}
		days = n
	}
	out, err := h.app.AnalyticsService().Upcoming(c.Request.Context(), uid, days)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	utils.WriteJSON(c.Writer, http.StatusOK, out)
}
