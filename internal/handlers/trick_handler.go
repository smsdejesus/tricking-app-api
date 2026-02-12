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
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"tricking-api/internal/services"
)

// TrickHandler handles HTTP requests for trick endpoints
type TrickHandler struct {
	// Depend on interface, not concrete type (enables testing with mocks)
	trickService services.TrickServiceInterface
}

// NewTrickHandler creates a new TrickHandler instance
func NewTrickHandler(trickService services.TrickServiceInterface) *TrickHandler {
	return &TrickHandler{trickService: trickService}
}

// GetSimpleTricksList returns a simple list of all tricks
func (h *TrickHandler) GetSimpleTricksList(c *gin.Context) {
	// Step 1: Get last modified timestamp from database (fast query)
	lastModified, err := h.trickService.GetLastModified(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve tricks",
		})
		return
	}

	// Step 2: Generate ETag from timestamp
	// Using timestamp-based ETag means we don't need to fetch/marshal data
	etag := fmt.Sprintf(`"%d"`, lastModified)

	// Step 3: Check If-None-Match header BEFORE fetching data
	// This is the key performance improvement - avoid expensive operations
	if c.GetHeader("If-None-Match") == etag {
		// Data hasn't changed, return 304 Not Modified
		c.Header("ETag", etag)
		c.Status(http.StatusNotModified)
		return
	}

	// Step 4: Only fetch data if ETag doesn't match (data has changed)
	tricks, err := h.trickService.GetSimpleTricksList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve tricks",
		})
		return
	}

	// Step 5: Build response
	responseData := gin.H{
		"tricks": tricks,
		"count":  len(tricks),
	}

	// Step 6: Set cache headers
	// public: can be cached by browsers and CDNs
	// max-age=3600: cache for 1 hour (3600 seconds)
	// stale-while-revalidate=86400: can serve stale content for 1 day while revalidating
	c.Header("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
	c.Header("ETag", etag)

	// Return successful response
	c.JSON(http.StatusOK, responseData)
}

// GetSimpleTrickById returns basic trick details
func (h *TrickHandler) GetSimpleTrickById(c *gin.Context) {
	// Parse ID from URL parameter
	id := c.Param("id")

	// Step 1: Get last modified timestamp for this specific trick
	lastModified, err := h.trickService.GetLastModifiedByID(c.Request.Context(), id)
	if err != nil {
		// Check for specific error types to return appropriate status codes
		if errors.Is(err, services.ErrTrickNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Trick not found",
			})
			return
		}

		// For other errors, continue without caching
		// (could also return error here, but we choose to be resilient)
	} else {
		// Step 2: Generate ETag from timestamp
		etag := fmt.Sprintf(`"%d"`, lastModified)

		// Step 3: Check If-None-Match header BEFORE fetching full data
		if c.GetHeader("If-None-Match") == etag {
			c.Header("ETag", etag)
			c.Status(http.StatusNotModified)
			return
		}

		// Set ETag for response
		c.Header("ETag", etag)
	}

	// Step 4: Fetch trick data (only if cache miss or ETag check failed)
	trick, err := h.trickService.GetSimpleTrickById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrTrickNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Trick not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve trick",
		})
		return
	}

	// Step 5: Set cache headers
	// Individual tricks change less frequently than lists, so longer cache
	c.Header("Cache-Control", "public, max-age=86400, stale-while-revalidate=604800")

	// Return response
	c.JSON(http.StatusOK, trick)
}

// GetFullDetailsTrickById returns full trick details with videos
func (h *TrickHandler) GetFullDetailsTrickById(c *gin.Context) {
	// Parse ID from URL parameter
	id := c.Param("id")

	// Step 1: Get last modified timestamp for this trick
	lastModified, err := h.trickService.GetLastModifiedByID(c.Request.Context(), id)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, services.ErrTrickNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Trick not found",
			})
			return
		}

		// For other errors, continue without caching
	} else {
		// Step 2: Generate ETag from timestamp
		etag := fmt.Sprintf(`"%d"`, lastModified)

		// Step 3: Check If-None-Match header BEFORE fetching data
		if c.GetHeader("If-None-Match") == etag {
			c.Header("ETag", etag)
			c.Status(http.StatusNotModified)
			return
		}

		// Set ETag for response
		c.Header("ETag", etag)
	}

	// Step 4: Fetch full trick details with videos
	trick, err := h.trickService.GetFullDetailsTrickById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrTrickNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Trick not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve trick details",
		})
		return
	}

	// Step 5: Set cache headers
	// Full details with videos - moderate cache duration
	c.Header("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")

	// Return response
	c.JSON(http.StatusOK, trick)
}
