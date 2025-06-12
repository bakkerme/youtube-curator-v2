package handlers

import (
	"context"

	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/email"
	"youtube-curator-v2/internal/processor"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
	"youtube-curator-v2/internal/summary"
	"youtube-curator-v2/internal/ytdlp"
)

// BaseHandlers contains shared dependencies and utilities for all API handlers
type BaseHandlers struct {
	store          store.Store
	feedProvider   rss.FeedProvider
	emailSender    email.Sender
	config         *config.Config
	processor      processor.ChannelProcessor
	videoStore     *store.VideoStore
	ytdlpEnricher  ytdlp.Enricher
	summaryService summary.SummaryServiceInterface
}

// NewBaseHandlers creates a new instance of base API handlers with all dependencies
func NewBaseHandlers(store store.Store, feedProvider rss.FeedProvider, emailSender email.Sender, cfg *config.Config, processor processor.ChannelProcessor, videoStore *store.VideoStore, ytdlpEnricher ytdlp.Enricher, summaryService summary.SummaryServiceInterface) *BaseHandlers {
	return &BaseHandlers{
		store:          store,
		feedProvider:   feedProvider,
		emailSender:    emailSender,
		config:         cfg,
		processor:      processor,
		videoStore:     videoStore,
		ytdlpEnricher:  ytdlpEnricher,
		summaryService: summaryService,
	}
}

// extractChannelIDWithYtdlpFallback tries ExtractChannelID first, then falls back to yt-dlp for unsupported URLs
func extractChannelIDWithYtdlpFallback(ctx context.Context, enricher ytdlp.Enricher, url string) (string, error) {
	return rss.ExtractChannelIDWithResolver(ctx, url, enricher)
}