package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"waterbilling/backend/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header required",
				"error":   "missing_token",
			})
			c.Abort()
			return
		}

		// Check if it's Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization format",
				"error":   "invalid_format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
				"error":   "invalid_token",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Not authenticated",
				"error":   "not_authenticated",
			})
			c.Abort()
			return
		}

		// Check if user has one of the allowed roles
		hasRole := false
		for _, role := range allowedRoles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Insufficient permissions",
				"error":   "insufficient_permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware logs requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log details
		duration := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()

		fmt.Printf("[%s] %s %s %d %v\n", clientIP, method, path, status, duration)
	}
}
