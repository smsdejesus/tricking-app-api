// =============================================================================
// FILE: internal/handlers/category_handler.go
// PURPOSE: HTTP request handling for category endpoints
// =============================================================================

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tricking-api/internal/services"
)

// CategoryHandler handles HTTP requests for category endpoints
type CategoryHandler struct {
	categoryService services.CategoryServiceInterface
}

// NewCategoryHandler creates a new CategoryHandler instance
func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// =============================================================================
// ENDPOINT: GET /categories
// PURPOSE: List all categories for UI dropdowns
// =============================================================================

// ListCategories returns all trick categories
// @Summary List all categories
// @Description Get all trick categories for filter dropdowns
// @Tags categories
// @Produce json
// @Success 200 {object} map[string]interface{} "categories array with count"
// @Failure 500 {object} map[string]string "Server error"
// @Router /categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryService.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"count":      len(categories),
	})
}
