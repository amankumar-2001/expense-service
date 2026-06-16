// Package autopay holds the service-to-service /v1/internal/autopays* HTTP
// handlers. They mirror the user-facing /v1/autopays* endpoints but take the
// operating user from a `userId` query parameter instead of a JWT — the whole
// group sits behind the shared-internal-key guard (ServiceAuth). The mcp-gateway
// is the intended caller: it acts on behalf of the Gmail-linked user it was
// configured with.
package autopay

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/di"
	"github.com/kharchibook/expense-service/utils"
)

// Handler serves the /v1/internal/autopays* endpoints.
type Handler struct {
	app di.AppInterface
}

// NewHandler constructs the internal autopay handler.
func NewHandler(app di.AppInterface) *Handler {
	return &Handler{app: app}
}

// Routes mounts the internal autopay routes onto the given (ServiceAuth-guarded)
// router group.
func (h *Handler) Routes(r gin.IRouter) {
	r.GET("/autopays", h.List)
	r.POST("/autopays", h.CreateDetected)
	r.PATCH("/autopays/:id", h.Update)
	r.POST("/autopays/:id/confirm", h.Confirm)
}

// userID parses the required `userId` query parameter. On failure it writes a 400
// and returns false.
func userID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Query("userId"), 10, 64)
	if err != nil || id <= 0 {
		utils.WriteError(c.Writer, apperrors.BadRequestError("userId query parameter is required"))
		return 0, false
	}
	return id, true
}

// pathID parses the :id path parameter. On failure it writes a 400 and returns false.
func pathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		utils.WriteError(c.Writer, apperrors.BadRequestError("invalid id"))
		return 0, false
	}
	return id, true
}

func writeOK(c *gin.Context, v any) {
	utils.WriteJSON(c.Writer, http.StatusOK, v)
}
