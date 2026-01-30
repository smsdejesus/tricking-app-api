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

// ListCategories returns all trick categories
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
