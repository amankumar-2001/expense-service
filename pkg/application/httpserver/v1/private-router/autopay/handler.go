// Package autopay holds the authenticated /v1/autopays* HTTP handlers. The
// controller (struct + Routes) lives here; each endpoint method has its own file.
package autopay

import (
	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/di"
)

// Handler serves the /v1/autopays* endpoints.
type Handler struct {
	app di.AppInterface
}

// NewHandler constructs the autopay handler.
func NewHandler(app di.AppInterface) *Handler {
	return &Handler{app: app}
}

// Routes mounts the autopay routes onto the given router group.
func (h *Handler) Routes(r gin.IRouter) {
	r.GET("/autopays", h.List)
	r.POST("/autopays", h.Create)
	r.PATCH("/autopays/:id", h.Update)
	r.DELETE("/autopays/:id", h.Delete)
	r.POST("/autopays/:id/confirm", h.Confirm)
}
