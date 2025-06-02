package api

import (
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRouter creates and configures the Echo router with all API endpoints
func SetupRouter(store store.Store, feedProvider rss.FeedProvider) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Create handlers
	handlers := NewHandlers(store, feedProvider)

	// API routes
	api := e.Group("/api")

	// Channel management endpoints
	api.GET("/channels", handlers.GetChannels)
	api.POST("/channels", handlers.AddChannel)
	api.POST("/channels/import", handlers.ImportChannels)
	api.DELETE("/channels/:id", handlers.RemoveChannel)

	// Configuration endpoints
	api.GET("/config/interval", handlers.GetCheckInterval)
	api.PUT("/config/interval", handlers.SetCheckInterval)

	return e
}
