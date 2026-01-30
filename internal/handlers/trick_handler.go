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

// TrickHandler handles HTTP requests for trick endpoints
type TrickHandler struct {
	// Depend on interface, not concrete type (enables testing with mocks)
	trickService services.TrickServiceInterface
}

// NewTrickHandler creates a new TrickHandler instance
func NewTrickHandler(trickService *services.TrickService) *TrickHandler {
	return &TrickHandler{trickService: trickService}
}

// ListTricks returns a simple list of all tricks
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

// GetTrickSimple returns basic trick details
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

// GetTrickFullDetails returns full trick details with videos

func (h *TrickHandler) GetTrickFullDetails(c *gin.Context) {
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
	trick, err := h.trickService.GetTrickFullDetails(c.Request.Context(), id)
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
