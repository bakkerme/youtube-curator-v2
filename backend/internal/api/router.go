package api

import (
	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/email"
	"youtube-curator-v2/internal/processor"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
	"youtube-curator-v2/internal/ytdlp"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRouter creates and configures the Echo router with all API endpoints
func SetupRouter(store store.Store, feedProvider rss.FeedProvider, emailSender email.Sender, cfg *config.Config, channelProcessor processor.ChannelProcessor, videoStore *store.VideoStore, ytdlpEnricher ytdlp.Enricher) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Create handlers
	handlers := NewHandlers(store, feedProvider, emailSender, cfg, channelProcessor, videoStore, ytdlpEnricher)

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
	api.GET("/config/smtp", handlers.GetSMTPConfig)
	api.PUT("/config/smtp", handlers.SetSMTPConfig)

	// Newsletter endpoints
	api.POST("/newsletter/run", handlers.RunNewsletter)

	// Video endpoints
	api.GET("/videos", handlers.GetVideos)
	api.POST("/videos/:videoId/watch", handlers.MarkVideoAsWatched)

	return e
}
