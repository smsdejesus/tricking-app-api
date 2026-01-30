package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"tricking-api/internal/models"
	"tricking-api/internal/services"
)

// ComboHandler handles HTTP requests for combo endpoints
type ComboHandler struct {
	comboService services.ComboServiceInterface
}

// NewComboHandler creates a new ComboHandler instance
func NewComboHandler(comboService *services.ComboService) *ComboHandler {
	return &ComboHandler{comboService: comboService}
}

// GenerateComboWithFilters creates a new random combo based on filters
func (h *ComboHandler) GenerateComboWithFilters(c *gin.Context) {
	var req models.ComboGenerateRequest

	// ShouldBindQuery also performs validation based on `binding` struct tags
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request parameters",
			// Include validation details in development, hide in production
			"details": err.Error(),
		})
		return
	}
	// Generate the combo
	combo, err := h.comboService.GenerateComboWithFilters(c.Request.Context(), req)
	if err != nil {
		// Check for specific errors
		if errors.Is(err, services.ErrInsufficientTricks) {
			// 422 Unprocessable Entity - request is valid but can't be fulfilled
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": err.Error(),
			})
			return
		}

		if errors.Is(err, services.ErrInvalidComboSize) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate combo",
		})
		return
	}

	c.JSON(http.StatusOK, combo)
}

// GenerateSimpleCombo creates a new random combo based only on size
func (h *ComboHandler) GenerateSimpleCombo(c *gin.Context) {
	// ==========================================================================
	// PARSE SINGLE QUERY PARAMETER
	// ==========================================================================
	// For simple cases, you can use ShouldBindQuery with a small struct
	// or parse individual parameters directly

	var req models.ComboGenerateSimpleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": "size is required and must be between 1 and 10",
		})
		return
	}

	// ==========================================================================
	// ALTERNATIVE: Manual parsing (for reference)
	// ==========================================================================
	// sizeStr := c.Query("size") // Returns empty string if not present
	// sizeStr := c.DefaultQuery("size", "3") // Returns "3" if not present
	//
	// size, err := strconv.Atoi(sizeStr)
	// if err != nil || size < 1 || size > 10 {
	//     c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid size"})
	//     return
	// }

	combo, err := h.comboService.GenerateSimpleCombo(c.Request.Context(), req.Size)
	if err != nil {
		if errors.Is(err, services.ErrInsufficientTricks) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": err.Error(),
			})
			return
		}

		if errors.Is(err, services.ErrInvalidComboSize) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate combo",
		})
		return
	}

	c.JSON(http.StatusOK, combo)
}
