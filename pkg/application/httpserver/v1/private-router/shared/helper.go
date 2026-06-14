// Package shared holds helpers reused across the authenticated v1 handler
// packages (expense, autopay, analytics).
package shared

import (
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/utils"
)

// UserID returns the authenticated user's ID, or (0, false) if absent. Handlers
// behind the JWT guard can treat false as an internal error; the 401 response is
// already written.
func UserID(c *gin.Context) (int64, bool) {
	id := utils.GetUserIDFromContext(c.Request.Context())
	if id == 0 {
		utils.WriteError(c.Writer, apperrors.UnauthorizedError("unauthenticated"))
		return 0, false
	}
	return id, true
}

// PathID parses an :id path parameter as int64. On failure it writes a 400 and
// returns false.
func PathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		utils.WriteError(c.Writer, apperrors.BadRequestError("invalid id"))
		return 0, false
	}
	return id, true
}
