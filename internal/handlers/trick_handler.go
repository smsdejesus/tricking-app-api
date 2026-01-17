// =============================================================================
// FILE: internal/handlers/trick_handler.go
// PURPOSE: HTTP request handling for trick endpoints
// =============================================================================
//
// HANDLER LAYER:
// Handlers are the bridge between HTTP and your application logic.
// They handle:
// - Parsing request data (URL params, query strings, JSON bodies)
// - Input validation (Gin provides binding validation)
// - Calling service methods
// - Formatting HTTP responses (status codes, JSON)
// - Error handling and error responses
//
// NAMING CONVENTIONS:
// - Handler suffix: TrickHandler, ComboHandler, etc.
// - Method names often match HTTP verbs + resource: GetTrick, ListTricks, CreateTrick
// - Methods must have signature: func(c *gin.Context)
//
// GIN CONTEXT:
// *gin.Context is your HTTP request/response object. It provides:
// - c.Param("name") - URL path parameters like /tricks/:id
// - c.Query("name") - Query string parameters like ?page=1
// - c.ShouldBindJSON(&obj) - Parse JSON body into struct
// - c.ShouldBindQuery(&obj) - Parse query params into struct
// - c.JSON(status, obj) - Send JSON response
// - c.AbortWithStatusJSON(status, obj) - Send error and stop processing
// =============================================================================

package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tricking-api/internal/services"
)

// =============================================================================
// HANDLER STRUCT
// =============================================================================

// TrickHandler handles HTTP requests for trick endpoints
type TrickHandler struct {
	// Depend on interface, not concrete type (enables testing with mocks)
	trickService services.TrickServiceInterface
}

// NewTrickHandler creates a new TrickHandler instance
func NewTrickHandler(trickService *services.TrickService) *TrickHandler {
	return &TrickHandler{trickService: trickService}
}

// =============================================================================
// ENDPOINT: GET /tricks
// PURPOSE: List all tricks (minimal data for dropdowns)
// =============================================================================

// ListTricks returns a simple list of all tricks
// @Summary List all tricks
// @Description Get a minimal list of tricks for dropdown menus
// @Tags tricks
// @Produce json
// @Success 200 {array} models.TrickSimpleResponse
// @Router /tricks [get]
func (h *TrickHandler) ListTricks(c *gin.Context) {
	// Call service method
	tricks, err := h.trickService.GetTricksList(c.Request.Context())
	if err != nil {
		// Log the error (in production, use a proper logger)
		// log.Printf("Error listing tricks: %v", err)

		// Return generic error to client (don't expose internal details)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve tricks",
		})
		return
	}

	// Return successful response
	// gin.H is a shortcut for map[string]interface{}
	c.JSON(http.StatusOK, gin.H{
		"tricks": tricks,
		"count":  len(tricks),
	})
}

// =============================================================================
// ENDPOINT: GET /tricks/:id
// PURPOSE: Get simple trick details (no videos)
// =============================================================================

// GetTrickSimple returns basic trick details
// @Summary Get trick details (simple)
// @Description Get basic trick information without videos
// @Tags tricks
// @Produce json
// @Param id path int true "Trick ID"
// @Success 200 {object} models.TrickDetailResponse
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Trick not found"
// @Router /tricks/{id} [get]
func (h *TrickHandler) GetTrickSimple(c *gin.Context) {
	// ==========================================================================
	// PARSE URL PARAMETER
	// ==========================================================================
	// c.Param("id") gets the :id from the URL path /tricks/:id
	// The parameter name "id" MUST match what's defined in the route
	idStr := c.Param("id")

	// Convert string to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid trick ID - must be a number",
		})
		return
	}

	// ==========================================================================
	// CALL SERVICE
	// ==========================================================================
	trick, err := h.trickService.GetTrickSimple(c.Request.Context(), id)
	if err != nil {
		// Check for specific error types to return appropriate status codes
		if errors.Is(err, services.ErrTrickNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Trick not found",
			})
			return
		}

		// Unexpected error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve trick",
		})
		return
	}

	// ==========================================================================
	// RETURN RESPONSE
	// ==========================================================================
	c.JSON(http.StatusOK, trick)
}

// =============================================================================
// ENDPOINT: GET /tricks/:id/dictionary
// PURPOSE: Get full trick details including videos
// =============================================================================

// GetTrickDictionary returns full trick details with videos
// @Summary Get trick dictionary page
// @Description Get complete trick information including all videos
// @Tags tricks
// @Produce json
// @Param id path int true "Trick ID"
// @Success 200 {object} models.TrickDictionaryResponse
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Trick not found"
// @Router /tricks/{id}/dictionary [get]
func (h *TrickHandler) GetTrickDictionary(c *gin.Context) {
	// Parse ID (same as above)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid trick ID - must be a number",
		})
		return
	}

	// Call the dictionary service method (includes videos)
	trick, err := h.trickService.GetTrickDictionary(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrTrickNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Trick not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve trick dictionary",
		})
		return
	}

	c.JSON(http.StatusOK, trick)
}

// =============================================================================
// RESPONSE FORMAT NOTES
// =============================================================================
//
// You have flexibility in how you format responses. Common patterns:
//
// PATTERN 1: Wrap in data object (recommended for consistency)
// c.JSON(200, gin.H{
//     "data": trick,
//     "meta": gin.H{"timestamp": time.Now()},
// })
//
// PATTERN 2: Direct object (simpler, used above)
// c.JSON(200, trick)
//
// PATTERN 3: Envelope for lists (shown in ListTricks)
// c.JSON(200, gin.H{
//     "tricks": tricks,
//     "count": len(tricks),
//     "page": 1,
//     "total_pages": 5,
// })
//
// PATTERN 4: Standard error format
// c.JSON(400, gin.H{
//     "error": gin.H{
//         "code": "INVALID_ID",
//         "message": "Trick ID must be a positive integer",
//     },
// })
//
// Choose one pattern and use it consistently across your API!
// =============================================================================
