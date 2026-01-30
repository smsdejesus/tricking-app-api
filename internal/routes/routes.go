package routes

import (
	"github.com/gin-gonic/gin"

	"tricking-api/internal/config"
	"tricking-api/internal/handlers"
	"tricking-api/internal/middleware"
)

func NewRouter(
	cfg *config.Config,
	trickHandler *handlers.TrickHandler,
	comboHandler *handlers.ComboHandler,
	categoryHandler *handlers.CategoryHandler,
	userHandler *handlers.UserHandler,
) *gin.Engine {
	// CREATE ROUTER
	router := gin.Default()

	// API VERSION GROUP
	// Routes will be:
	// /api/v1/tricks
	// /api/v1/combos
	// /api/v1/categories
	v1 := router.Group("/api/v1")
	// All routes require internal API key

	// V1 ROUTES
	{
		// GET /api/v1/tricks - List all tricks (for dropdowns/search)
		v1.GET("/tricks", trickHandler.ListTricks)

		// ======================================================================
		// TRICK ROUTES
		// ======================================================================
		tricks := v1.Group("/trick")
		{

			// GET /api/v1/tricks/:id - Get simple trick details
			// :id is a URL parameter - any value in that position is captured
			// Example: /api/v1/tricks/5 -> id = "5"
			tricks.GET("/:id", trickHandler.GetTrickSimple)

			// GET /api/v1/tricks/:id/dictionary - Get full trick details with videos
			// Nested resource - the dictionary "belongs to" a specific trick
			tricks.GET("/detail/:id", trickHandler.GetTrickFullDetails)
		}

		// ======================================================================
		// COMBO ROUTES
		// ======================================================================
		combos := v1.Group("/combos")
		{
			// GET /api/v1/combos/generate - Generate combo with filters
			// Using GET because this is a read operation (no data created)
			// Filters are passed as query parameters
			combos.GET("/generate", comboHandler.GenerateComboWithFilters)

			// GET /api/v1/combos/generate/simple - Generate combo with size only
			combos.GET("/generate/simple", comboHandler.GenerateSimpleCombo)
		}

		// ======================================================================
		// CATEGORY ROUTES
		// ======================================================================
		categories := v1.Group("/categories")
		{
			// GET /api/v1/categories - List all categories
			categories.GET("", categoryHandler.ListCategories)
		}

		// ======================================================================
		// USER ROUTES (for saved combos) NOT IMPLEMENTED YET
		// ======================================================================
		// Extract user context from BFF headers for all /users routes
		v1.Use(middleware.ExtractUserContext())
		v1.Use(middleware.InternalAPIKey(cfg.InternalAPIKey))
		users := v1.Group("/users")
		{
			// GET /api/v1/users/:userId/combos - Get user's saved combos
			// This is a nested resource - combos belong to a user
			users.GET("/:userId/combos", userHandler.GetUserCombos)
		}
	}

	// ==========================================================================
	// HEALTH CHECK ROUTE
	// ==========================================================================
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	return router
}
