package middleware

import (
	"net/http"
	"strings"

	"lifeline_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware protects routes by validating the JWT token
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer token"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ParseToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		userIDFloat, ok1 := (*claims)["user_id"].(float64)
		role, ok2 := (*claims)["role"].(string)
		if !ok1 || !ok2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userID", uint(userIDFloat))
		c.Set("role", role)
		c.Next()
	}
}

// RequireRole restricts route access to specific roles (e.g. PATIENT, PHARMACY)
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		role, ok := roleVal.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role configuration"})
			c.Abort()
			return
		}

		allowed := false
		for _, r := range allowedRoles {
			if strings.ToUpper(role) == strings.ToUpper(r) {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden: insufficient role permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
