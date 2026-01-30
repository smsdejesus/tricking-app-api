// =============================================================================
// FILE: internal/handlers/user_handler.go
// PURPOSE: HTTP request handling for user-related endpoints
// =============================================================================

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"tricking-api/internal/services"
)

// UserHandler handles HTTP requests for user endpoints
type UserHandler struct {
	userService services.UserServiceInterface
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUserCombos returns all saved combos for a user
func (h *UserHandler) GetUserCombos(c *gin.Context) {
	// =========================================================================
	// PARSE USER ID FROM URL
	// =========================================================================
	// This is WHOSE combos we want to fetch
	requestedUserID := c.Param("userId")

	parsedRequestedID, err := uuid.Parse(requestedUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format - must be a valid UUID",
		})
		return
	}

	// =========================================================================
	// AUTHORIZATION CHECK
	// =========================================================================
	// Compare requested user vs authenticated user (from BFF header)
	authenticatedUserID, exists := c.Get("user_id")

	// If we have an authenticated user, verify they can access this resource
	if exists && authenticatedUserID != "" {
		// User can only view their own combos (unless admin)
		if authenticatedUserID != requestedUserID {
			userRole, _ := c.Get("user_role")
			if userRole != "admin" {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "You can only view your own combos",
				})
				return
			}
		}
	}

	// =========================================================================
	// FETCH COMBOS
	// =========================================================================
	combos, err := h.userService.GetUserCombos(c.Request.Context(), parsedRequestedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve combos",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"combos": combos,
		"count":  len(combos),
	})
}
