package routes

import (
	"net/http"

	"lifeline_backend/internal/app"
	"lifeline_backend/internal/handlers"
	"lifeline_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Register wires ALL routes onto the app router.
func Register(a *app.App) {
	api := a.Router.Group("/api")
	v1 := api.Group("/v1")

	// System/probe routes under /api/v1
	registerSystemRoutes(v1, a)

	v1.GET("/health", healthHandler)

	// Auth routes
	authHandler := handlers.NewAuthHandler(a.DB, a.Config.JWTSecret)
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/signup", authHandler.Signup)
		authGroup.POST("/login", authHandler.Login)
	}

	// Protected routes group
	authMiddleware := middleware.AuthMiddleware(a.Config.JWTSecret)
	protected := v1.Group("/")
	protected.Use(authMiddleware)
	{
		// Patient only routes
		patientOnly := protected.Group("/")
		patientOnly.Use(middleware.RequireRole("PATIENT"))
		{
			requestHandler := handlers.NewRequestHandler(a.DB, a.Hub)
			patientOnly.POST("/requests", requestHandler.CreateRequest)
			patientOnly.POST("/requests/:id/close", requestHandler.CloseRequest)
			patientOnly.GET("/requests", requestHandler.ListPatientRequests)
		}

		// Pharmacy only routes
		pharmacyOnly := protected.Group("/")
		pharmacyOnly.Use(middleware.RequireRole("PHARMACY"))
		{
			responseHandler := handlers.NewResponseHandler(a.DB, a.Hub)
			pharmacyOnly.POST("/responses", responseHandler.CreateResponse)

			requestHandler := handlers.NewRequestHandler(a.DB, a.Hub)
			pharmacyOnly.GET("/pharmacy/requests", requestHandler.ListPharmacyIncomingRequests)
		}

		// General protected routes
		responseHandler := handlers.NewResponseHandler(a.DB, a.Hub)
		protected.GET("/requests/:id/responses", responseHandler.ListResponsesForRequest)

		wsHandler := handlers.NewWSHandler(a.Hub, a.Config.JWTSecret)
		protected.GET("/requests/:id/messages", wsHandler.GetChatMessages)
	}

	// WebSocket route for pharmacy alerts (authenticates via query param internally)
	wsHandler := handlers.NewWSHandler(a.Hub, a.Config.JWTSecret)
	v1.GET("/ws/pharmacy", wsHandler.ConnectPharmacy)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "server running",
	})
}
