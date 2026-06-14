// Package analytics holds the authenticated free-money HTTP handlers: the salary
// input and the derived committed / upcoming analytics. The controller (struct +
// Routes) lives here; each endpoint method has its own file.
package analytics

import (
	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/pkg/di"
)

// Handler serves the salary + /v1/analytics* endpoints.
type Handler struct {
	app di.AppInterface
}

// NewHandler constructs the analytics handler.
func NewHandler(app di.AppInterface) *Handler {
	return &Handler{app: app}
}

// defaultUpcomingDays is the on-demand window for upcoming deductions.
const defaultUpcomingDays = 7

// Routes mounts the salary + analytics routes onto the given router group.
func (h *Handler) Routes(r gin.IRouter) {
	r.PUT("/salary", h.SetSalary)
	r.GET("/analytics/committed", h.Committed)
	r.GET("/analytics/upcoming", h.Upcoming)
}
