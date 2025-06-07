package api

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"youtube-curator-v2/internal/config"
	"youtube-curator-v2/internal/email"
	"youtube-curator-v2/internal/processor"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"github.com/labstack/echo/v4"
)

// Handlers contains the API handlers
type Handlers struct {
	store        store.Store
	feedProvider rss.FeedProvider
	emailSender  email.Sender
	config       *config.Config
	processor    processor.ChannelProcessor
	videoStore   *store.VideoStore
}

// NewHandlers creates a new instance of API handlers
func NewHandlers(store store.Store, feedProvider rss.FeedProvider, emailSender email.Sender, cfg *config.Config, processor processor.ChannelProcessor, videoStore *store.VideoStore) *Handlers {
	return &Handlers{
		store:        store,
		feedProvider: feedProvider,
		emailSender:  emailSender,
		config:       cfg,
		processor:    processor,
		videoStore:   videoStore,
	}
}

// Channel represents a channel in API responses
type Channel struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ChannelRequest represents a request to add a channel
// Title is optional; if not provided, it will be fetched from RSS
type ChannelRequest struct {
	URL   string `json:"url" validate:"required"`
	Title string `json:"title,omitempty"`
}

// ConfigInterval represents the check interval configuration
type ConfigInterval struct {
	Interval string `json:"interval"`
}

// SMTPConfigRequest represents a request to update SMTP configuration
type SMTPConfigRequest struct {
	Server         string `json:"server" validate:"required"`
	Port           string `json:"port" validate:"required"`
	Username       string `json:"username" validate:"required"`
	Password       string `json:"password" validate:"required"`
	RecipientEmail string `json:"recipientEmail" validate:"required,email"`
}

// SMTPConfigResponse represents SMTP configuration in API responses (without password)
type SMTPConfigResponse struct {
	Server         string `json:"server"`
	Port           string `json:"port"`
	Username       string `json:"username"`
	RecipientEmail string `json:"recipientEmail"`
	PasswordSet    bool   `json:"passwordSet"`
}

// ImportChannelsRequest represents a request to import multiple channels
type ImportChannelsRequest struct {
	Channels []ChannelImport `json:"channels" validate:"required"`
}

// ChannelImport represents a channel to be imported
type ChannelImport struct {
	URL   string `json:"url" validate:"required"`
	Title string `json:"title,omitempty"`
}

// ImportChannelsResponse represents the response from importing channels
type ImportChannelsResponse struct {
	Imported []Channel       `json:"imported"`
	Failed   []ImportFailure `json:"failed"`
}

// ImportFailure represents a failed channel import
type ImportFailure struct {
	Channel ChannelImport `json:"channel"`
	Error   string        `json:"error"`
}

// GetChannels handles GET /api/channels
func (h *Handlers) GetChannels(c echo.Context) error {
	channels, err := h.store.GetChannels()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve channels")
	}

	if channels == nil {
		channels = []store.Channel{}
	}

	return c.JSON(http.StatusOK, channels)
}

// AddChannel handles POST /api/channels
func (h *Handlers) AddChannel(c echo.Context) error {
	var req ChannelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.URL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "URL is required")
	}

	// Extract channel ID from URL
	channelID, err := rss.ExtractChannelID(req.URL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	title := req.Title
	if title == "" {
		// Fetch title from RSS feed
		ctx := context.Background()
		feed, err := h.feedProvider.FetchFeed(ctx, channelID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not fetch channel title from RSS feed: "+err.Error())
		}
		title = feed.Title
		if title == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Channel title could not be determined from RSS feed")
		}
	}

	channel := store.Channel{ID: channelID, Title: title}
	if err := h.store.AddChannel(channel); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add channel")
	}

	return c.JSON(http.StatusCreated, channel)
}

// RemoveChannel handles DELETE /api/channels/:id
func (h *Handlers) RemoveChannel(c echo.Context) error {
	channelID := c.Param("id")
	if channelID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Channel ID is required")
	}

	// Validate channel ID format
	if err := rss.ValidateChannelID(channelID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Remove channel from store
	if err := h.store.RemoveChannel(channelID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to remove channel")
	}

	return c.NoContent(http.StatusNoContent)
}

// GetCheckInterval handles GET /api/config/interval
func (h *Handlers) GetCheckInterval(c echo.Context) error {
	interval, err := h.store.GetCheckInterval()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve check interval")
	}

	return c.JSON(http.StatusOK, ConfigInterval{Interval: interval.String()})
}

// SetCheckInterval handles PUT /api/config/interval
func (h *Handlers) SetCheckInterval(c echo.Context) error {
	var req ConfigInterval
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Interval == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Interval is required")
	}

	// Parse duration
	duration, err := time.ParseDuration(req.Interval)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid interval format. Use Go duration format (e.g., '1h', '30m', '2h30m')")
	}

	// Validate reasonable range (1 minute to 24 hours)
	if duration < time.Minute {
		return echo.NewHTTPError(http.StatusBadRequest, "Interval must be at least 1 minute")
	}
	if duration > 24*time.Hour {
		return echo.NewHTTPError(http.StatusBadRequest, "Interval must be no more than 24 hours")
	}

	// Set interval in store
	if err := h.store.SetCheckInterval(duration); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set check interval")
	}

	return c.JSON(http.StatusOK, ConfigInterval{Interval: duration.String()})
}

// GetSMTPConfig handles GET /api/config/smtp
func (h *Handlers) GetSMTPConfig(c echo.Context) error {
	config, err := h.store.GetSMTPConfig()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve SMTP configuration")
	}

	// If no config exists, return empty response
	if config == nil {
		return c.JSON(http.StatusOK, SMTPConfigResponse{
			PasswordSet: false,
		})
	}

	// Return config without password
	response := SMTPConfigResponse{
		Server:         config.Server,
		Port:           config.Port,
		Username:       config.Username,
		RecipientEmail: config.RecipientEmail,
		PasswordSet:    config.Password != "",
	}

	return c.JSON(http.StatusOK, response)
}

// SetSMTPConfig handles PUT /api/config/smtp
func (h *Handlers) SetSMTPConfig(c echo.Context) error {
	var req SMTPConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.Server == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server is required")
	}
	if req.Port == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Port is required")
	}
	if req.Username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Username is required")
	}
	if req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Password is required")
	}
	if req.RecipientEmail == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Recipient email is required")
	}

	// Basic email validation
	if !strings.Contains(req.RecipientEmail, "@") {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid recipient email format")
	}

	// Create SMTP config
	smtpConfig := &store.SMTPConfig{
		Server:         req.Server,
		Port:           req.Port,
		Username:       req.Username,
		Password:       req.Password,
		RecipientEmail: req.RecipientEmail,
	}

	// Save to store
	if err := h.store.SetSMTPConfig(smtpConfig); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save SMTP configuration")
	}

	// Return response without password
	response := SMTPConfigResponse{
		Server:         req.Server,
		Port:           req.Port,
		Username:       req.Username,
		RecipientEmail: req.RecipientEmail,
		PasswordSet:    true,
	}

	return c.JSON(http.StatusOK, response)
}

// ImportChannels handles POST /api/channels/import
func (h *Handlers) ImportChannels(c echo.Context) error {
	var req ImportChannelsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if len(req.Channels) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "At least one channel is required")
	}

	var imported []Channel
	var failed []ImportFailure

	ctx := context.Background()

	for _, channelImport := range req.Channels {
		if channelImport.URL == "" {
			failed = append(failed, ImportFailure{
				Channel: channelImport,
				Error:   "URL is required",
			})
			continue
		}

		// Extract channel ID from URL
		channelID, err := rss.ExtractChannelID(channelImport.URL)
		if err != nil {
			failed = append(failed, ImportFailure{
				Channel: channelImport,
				Error:   err.Error(),
			})
			continue
		}

		title := channelImport.Title
		if title == "" {
			// Fetch title from RSS feed
			feed, err := h.feedProvider.FetchFeed(ctx, channelID)
			if err != nil {
				failed = append(failed, ImportFailure{
					Channel: channelImport,
					Error:   "Could not fetch channel title from RSS feed: " + err.Error(),
				})
				continue
			}
			title = feed.Title
			if title == "" {
				failed = append(failed, ImportFailure{
					Channel: channelImport,
					Error:   "Channel title could not be determined from RSS feed",
				})
				continue
			}
		}

		channel := store.Channel{ID: channelID, Title: title}
		if err := h.store.AddChannel(channel); err != nil {
			failed = append(failed, ImportFailure{
				Channel: channelImport,
				Error:   "Failed to add channel to database: " + err.Error(),
			})
			continue
		}

		imported = append(imported, Channel{ID: channelID, Title: title})
	}

	response := ImportChannelsResponse{
		Imported: imported,
		Failed:   failed,
	}

	// Return 207 Multi-Status if there were any failures, 201 if all succeeded
	statusCode := http.StatusCreated
	if len(failed) > 0 {
		statusCode = http.StatusMultiStatus
	}

	return c.JSON(statusCode, response)
}

// RunNewsletterRequest represents a request to manually trigger newsletter run
type RunNewsletterRequest struct {
	ChannelID         string `json:"channelId,omitempty"`
	IgnoreLastChecked bool   `json:"ignoreLastChecked,omitempty"`
	MaxItems          int    `json:"maxItems,omitempty"`
}

// RunNewsletter handles POST /api/newsletter/run
func (h *Handlers) RunNewsletter(c echo.Context) error {
	var req RunNewsletterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// If channelID is provided, validate it
	if req.ChannelID != "" {
		if err := rss.ValidateChannelID(req.ChannelID); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	// Validate maxItems if provided
	if req.MaxItems < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "maxItems must be non-negative")
	}

	ctx := context.Background()
	var channels []store.Channel
	var err error

	if req.ChannelID != "" {
		// Get specific channel
		allChannels, err := h.store.GetChannels()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve channels")
		}

		// Find the specific channel
		found := false
		for _, ch := range allChannels {
			if ch.ID == req.ChannelID {
				channels = append(channels, ch)
				found = true
				break
			}
		}

		if !found {
			return echo.NewHTTPError(http.StatusBadRequest, "Channel not found")
		}
	} else {
		// Get all channels
		channels, err = h.store.GetChannels()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve channels")
		}
	}

	if len(channels) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No channels configured")
	}

	// Process channels and collect new videos
	var allNewVideos []rss.Entry
	processedCount := 0
	errorCount := 0

	for _, channel := range channels {
		result := h.processor.ProcessChannelWithOptions(ctx, channel.ID, req.IgnoreLastChecked, req.MaxItems)

		if result.Error != nil {
			errorCount++
			continue
		}

		processedCount++
		if result.NewVideo != nil {
			allNewVideos = append(allNewVideos, *result.NewVideo)
		}
	}

	// Sort videos by published date
	sort.Slice(allNewVideos, func(i, j int) bool {
		return allNewVideos[i].Published.Before(allNewVideos[j].Published)
	})

	// Send email if there are new videos
	if len(allNewVideos) > 0 {
		emailBody, err := email.FormatNewVideosEmail(allNewVideos)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to format email: "+err.Error())
		}

		// Get SMTP config from database
		smtpConfig, err := h.store.GetSMTPConfig()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve SMTP configuration: "+err.Error())
		}
		if smtpConfig == nil || smtpConfig.RecipientEmail == "" {
			return echo.NewHTTPError(http.StatusInternalServerError, "SMTP configuration not set")
		}

		subject := "New YouTube Videos Update"
		if err := h.emailSender.Send(smtpConfig.RecipientEmail, subject, emailBody); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send email: "+err.Error())
		}
	}

	// Return response with stats
	response := map[string]interface{}{
		"message":           "Newsletter run completed",
		"channelsProcessed": processedCount,
		"channelsWithError": errorCount,
		"newVideosFound":    len(allNewVideos),
		"emailSent":         len(allNewVideos) > 0,
	}

	return c.JSON(http.StatusOK, response)
}

// GetVideos handles GET /api/videos - returns all videos from the video store
func (h *Handlers) GetVideos(c echo.Context) error {
	if h.videoStore == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Video store not initialized")
	}

	// Check for refresh parameter
	refresh := c.QueryParam("refresh") == "true"

	// Get current cached videos
	videos := h.videoStore.GetAllVideos()

	// If no cached videos or refresh requested, fetch from channels
	if len(videos) == 0 || refresh {
		ctx := context.Background()

		// Get all channels
		channels, err := h.store.GetChannels()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve channels")
		}

		// Fetch most recent video from each channel
		for _, channel := range channels {
			feed, err := h.feedProvider.FetchFeed(ctx, channel.ID)
			if err != nil {
				// Continue with other channels if one fails
				continue
			}

			// Get the most recent video (first entry)
			if len(feed.Entries) > 0 {
				mostRecentVideo := feed.Entries[0]
				h.videoStore.AddVideo(channel.ID, mostRecentVideo)
			}
		}

		// Get updated videos after fetching
		videos = h.videoStore.GetAllVideos()
	}

	// Sort videos by published date (newest first)
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Entry.Published.After(videos[j].Entry.Published)
	})

	return c.JSON(http.StatusOK, videos)
}
