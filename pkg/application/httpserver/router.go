// Package httpserver builds the application's HTTP handler: the global
// middleware chain plus the per-version route mounts.
package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/middleware"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/analytics"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/autopay"
	"github.com/kharchibook/expense-service/pkg/application/httpserver/v1/private-router/expense"
	"github.com/kharchibook/expense-service/pkg/di"
	"github.com/kharchibook/expense-service/utils"
)

// NewRouter assembles the HTTP handler from the application container.
func NewRouter(app di.AppInterface) http.Handler {
	// Quiet, production-friendly mode unless running the dev env.
	if app.Config().App.Env == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Reject request bodies with unknown fields (matches the prior strict decoder).
	gin.EnableJsonDecoderDisallowUnknownFields()

	r := gin.New()

	// Global middleware chain (outermost first). RequestInfo derives the client
	// IP from X-Forwarded-For/X-Real-Ip for logging.
	r.Use(middleware.RequestInfo())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	guard := middleware.NewGuard(app.TokenService())

	// Liveness/readiness.
	r.GET("/healthz", func(c *gin.Context) {
		utils.WriteJSON(c.Writer, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if err := app.HealthCheck(c.Request.Context()); err != nil {
			utils.WriteJSON(c.Writer, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		utils.WriteJSON(c.Writer, http.StatusOK, map[string]string{"status": "ready"})
	})

	// V1 routes — every route requires a valid access token, so the guard is
	// applied to the whole group.
	mountV1(r.Group("/v1", guard.JWT), app)

	return r
}

// mountV1 wires the V1 authenticated routers, one per resource package.
func mountV1(r *gin.RouterGroup, app di.AppInterface) {
	expense.NewHandler(app).Routes(r)
	autopay.NewHandler(app).Routes(r)
	analytics.NewHandler(app).Routes(r)
}
