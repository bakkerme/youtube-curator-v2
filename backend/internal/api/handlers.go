package api

import (
	"context"
	"net/http"
	"time"

	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"github.com/labstack/echo/v4"
)

// Handlers contains the API handlers
type Handlers struct {
	store        store.Store
	feedProvider rss.FeedProvider
}

// NewHandlers creates a new instance of API handlers
func NewHandlers(store store.Store, feedProvider rss.FeedProvider) *Handlers {
	return &Handlers{
		store:        store,
		feedProvider: feedProvider,
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
