package routes

import (
	"context"
	"net/http"
	"time"

	"lifeline_backend/internal/app"

	"github.com/gin-gonic/gin"
)

// RouteParam documents a single query/path parameter.
type RouteParam struct {
	Name     string `json:"name"`
	In       string `json:"in"` // "path", "query", "body"
	Required bool   `json:"required"`
	Type     string `json:"type"` // "string", "int", "bool", etc.
	Desc     string `json:"description"`
}

// RouteDoc documents a single route for the API index.
type RouteDoc struct {
	Method      string       `json:"method"`
	Path        string       `json:"path"`
	Description string       `json:"description"`
	Params      []RouteParam `json:"params,omitempty"`
	Response    any          `json:"example_response"`
}

// apiIndex returns the live API reference served at GET /.
// Update this whenever routes are added or changed.
func apiIndex() gin.H {
	return gin.H{
		"service":  "LifeLine Backend API",
		"version":  "v1",
		"base_url": "/api/v1",
		"docs":     "All endpoints are prefixed with /api/v1",
		"routes": []RouteDoc{
			{
				Method:      "GET",
				Path:        "/api/v1/health",
				Description: "Check whether the server process is running.",
				Response:    gin.H{"success": true, "message": "server running"},
			},
			{
				Method:      "GET",
				Path:        "/api/v1/healthz",
				Description: "Liveness probe — confirms the process is alive (used by load balancers / k8s).",
				Response:    gin.H{"status": "ok"},
			},
			{
				Method:      "GET",
				Path:        "/api/v1/readyz",
				Description: "Readiness probe — confirms the server can reach the database and is ready to serve traffic.",
				Response:    gin.H{"status": "ready"},
			},
		},
	}
}

// registerSystemRoutes mounts infra/probe routes under the v1 group.
func registerSystemRoutes(v1 *gin.RouterGroup, a *app.App) {
	// Root — live API reference for frontend/dev teams
	a.Router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, apiIndex())
	})

	// Liveness probe — is the process alive?
	v1.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Readiness probe — is the process ready to serve traffic?
	v1.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := a.DB.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}

		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}
