// =============================================================================
// FILE: internal/handlers/combo_handler.go
// PURPOSE: HTTP request handling for combo endpoints
// =============================================================================

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

// =============================================================================
// ENDPOINT: GET /combos/generate
// PURPOSE: Generate a combo with filters
// =============================================================================

// GenerateCombo creates a new random combo based on filters
// @Summary Generate a combo with filters
// @Description Generate a random trick combo using provided filters
// @Tags combos
// @Accept json
// @Produce json
// @Param request query models.ComboGenerateRequest true "Generation parameters"
// @Success 200 {object} models.GeneratedComboResponse
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 422 {object} map[string]string "Not enough tricks available"
// @Router /combos/generate [get]
func (h *ComboHandler) GenerateCombo(c *gin.Context) {
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

	// ==========================================================================
	// CALL SERVICE
	// ==========================================================================
	combo, err := h.comboService.GenerateCombo(c.Request.Context(), req)
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

// =============================================================================
// ENDPOINT: GET /combos/generate/simple
// PURPOSE: Generate a combo with only size (no filters)
// =============================================================================

// GenerateSimpleCombo creates a new random combo based only on size
// @Summary Generate a simple combo
// @Description Generate a random trick combo using only size parameter
// @Tags combos
// @Produce json
// @Param size query int true "Number of tricks in combo" minimum(1) maximum(20)
// @Success 200 {object} models.GeneratedComboResponse
// @Failure 400 {object} map[string]string "Invalid size"
// @Failure 422 {object} map[string]string "Not enough tricks available"
// @Router /combos/generate/simple [get]
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
			"details": "size is required and must be between 1 and 20",
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
	// if err != nil || size < 1 || size > 20 {
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
