package middleware

import (
	"net/http"
	"strings"

	"belimang/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_authorization_header", "message": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_authorization_header", "message": "Authorization header must start with 'Bearer '"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		// Continue with the next handler
		c.Next()
	}
}

func RequireUserRole(jwtService *jwt.JWTService, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_authorization_header", "message": "Authorization header is required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_authorization_header", "message": "Authorization header must start with 'Bearer '"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims.Role != requiredRole {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "insufficient_permissions", "message": "You don't have permission to access this resource"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func RequireAdmin(jwtService *jwt.JWTService) gin.HandlerFunc {
	return RequireUserRole(jwtService, "admin")
}

func RequireUser(jwtService *jwt.JWTService) gin.HandlerFunc {
	return RequireUserRole(jwtService, "user")
}
