package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// InternalAPIKey validates that requests come from your BFF
// This is a simple approach - the BFF sends a secret API key
func InternalAPIKey(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("internal-api-key")

		if apiKey == "" || apiKey != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing API key",
			})
			return
		}

		c.Next()
	}
}

// ExtractUserContext pulls user info that the BFF passes in headers
// The BFF already authenticated the user - we just need their ID
func ExtractUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// BFF sends user info in headers after authenticating them
		userID := c.GetHeader("user-id")
		userRole := c.GetHeader("user-role")

		// Store in context for handlers to use
		if userID != "" {
			c.Set("user_id", userID)
		}
		if userRole != "" {
			c.Set("user_role", userRole)
		}

		c.Next()
	}
}
