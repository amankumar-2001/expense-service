package autopay

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/shared"
	"github.com/kharchibook/expense-service/utils"
)

// Delete soft-deletes (cancels) an autopay.
func (h *Handler) Delete(c *gin.Context) {
	uid, ok := shared.UserID(c)
	if !ok {
		return
	}
	id, ok := shared.PathID(c)
	if !ok {
		return
	}
	if err := h.app.AutoPayService().Delete(c.Request.Context(), uid, id); err != nil {
		utils.WriteError(c.Writer, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
