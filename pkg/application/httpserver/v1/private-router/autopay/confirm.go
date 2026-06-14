package autopay

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/utils"
)

// Confirm activates an auto-detected autopay.
func (h *Handler) Confirm(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	id, ok := shared.PathID(c)
	if !ok {
		return
	}
	out, err := h.app.AutoPayService().Confirm(c.Request.Context(), uid, id)
	if err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	utils.WriteJSON(c.Writer, http.StatusOK, out)
}
