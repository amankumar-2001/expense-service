// Package expense holds the authenticated /v1/expenses* HTTP handlers. The
// controller (struct + Routes) lives here; each endpoint method has its own file.
package expense

import (
	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/di"
)

// Handler serves the /v1/expenses* endpoints.
type Handler struct {
	app di.AppInterface
}

// NewHandler constructs the expense handler.
func NewHandler(app di.AppInterface) *Handler {
	return &Handler{app: app}
}

// Routes mounts the expense routes onto the given router group.
func (h *Handler) Routes(r gin.IRouter) {
	r.POST("/expenses", h.Create)
	r.GET("/expenses", h.List)
	r.DELETE("/expenses/last", h.DeleteLast)
	r.GET("/expenses/summary", h.Summary)
}
