package api

import (
	"youtube-curator-v2/internal/api/handlers"
	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/email"
	"youtube-curator-v2/internal/processor"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
	"youtube-curator-v2/internal/summary"
	"youtube-curator-v2/internal/ytdlp"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRouter creates and configures the Echo router with all API endpoints
func SetupRouter(store store.Store, feedProvider rss.FeedProvider, emailSender email.Sender, cfg *config.Config, channelProcessor processor.ChannelProcessor, videoStore *store.VideoStore, ytdlpEnricher ytdlp.Enricher, summaryService summary.SummaryServiceInterface) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Create base handlers with shared dependencies
	baseHandlers := handlers.NewBaseHandlers(store, feedProvider, emailSender, cfg, channelProcessor, videoStore, ytdlpEnricher, summaryService)

	// Create domain-specific handlers
	channelHandlers := handlers.NewChannelHandlers(baseHandlers)
	configHandlers := handlers.NewConfigHandlers(baseHandlers)
	videoHandlers := handlers.NewVideoHandlers(baseHandlers)
	newsletterHandlers := handlers.NewNewsletterHandlers(baseHandlers)

	// API routes
	api := e.Group("/api")

	// Channel management endpoints
	api.GET("/channels", channelHandlers.GetChannels)
	api.POST("/channels", channelHandlers.AddChannel)
	api.POST("/channels/import", channelHandlers.ImportChannels)
	api.DELETE("/channels/:id", channelHandlers.RemoveChannel)

	// Configuration endpoints
	api.GET("/config/interval", configHandlers.GetCheckInterval)
	api.PUT("/config/interval", configHandlers.SetCheckInterval)
	api.GET("/config/smtp", configHandlers.GetSMTPConfig)
	api.PUT("/config/smtp", configHandlers.SetSMTPConfig)
	api.GET("/config/llm", configHandlers.GetLLMConfig)
	api.PUT("/config/llm", configHandlers.SetLLMConfig)
	api.GET("/config/newsletter", configHandlers.GetNewsletterConfig)
	api.PUT("/config/newsletter", configHandlers.SetNewsletterConfig)

	// Newsletter endpoints
	api.POST("/newsletter/run", newsletterHandlers.RunNewsletter)

	// Video endpoints
	api.GET("/videos", videoHandlers.GetVideos)
	api.POST("/videos/:videoId/watch", videoHandlers.MarkVideoAsWatched)
	api.GET("/videos/:videoId/summary", videoHandlers.GetVideoSummary)

	return e
}
