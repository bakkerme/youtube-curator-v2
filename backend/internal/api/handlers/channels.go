package handlers

import (
	"context"
	"net/http"

	"youtube-curator-v2/internal/api/types"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"github.com/labstack/echo/v4"
)

// ChannelHandlers provides handlers for channel management endpoints
type ChannelHandlers struct {
	*BaseHandlers
}

// NewChannelHandlers creates a new instance of channel handlers
func NewChannelHandlers(base *BaseHandlers) *ChannelHandlers {
	return &ChannelHandlers{BaseHandlers: base}
}

// GetChannels handles GET /api/channels
func (h *ChannelHandlers) GetChannels(c echo.Context) error {
	channels, err := h.store.GetChannels()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve channels")
	}

	if channels == nil {
		channels = []store.Channel{}
	}

	response := types.TransformChannels(channels)
	return c.JSON(http.StatusOK, response)
}

// AddChannel handles POST /api/channels
func (h *ChannelHandlers) AddChannel(c echo.Context) error {
	var req types.ChannelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.URL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "URL is required")
	}

	// Extract channel ID from URL, with yt-dlp fallback for @username, /c/, /user/ URLs
	channelID, err := extractChannelIDWithYtdlpFallback(c.Request().Context(), h.ytdlpEnricher, req.URL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	title := req.Title
	if title == "" {
		// Fetch title from RSS feed
		ctx := c.Request().Context()
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

	response := types.TransformChannel(channel)
	return c.JSON(http.StatusCreated, response)
}

// RemoveChannel handles DELETE /api/channels/:id
func (h *ChannelHandlers) RemoveChannel(c echo.Context) error {
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

// ImportChannels handles POST /api/channels/import
func (h *ChannelHandlers) ImportChannels(c echo.Context) error {
	var req types.ImportChannelsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if len(req.Channels) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "At least one channel is required")
	}

	var imported []types.ChannelResponse
	var failed []types.ImportFailure

	ctx := context.Background()

	for _, channelImport := range req.Channels {
		if channelImport.URL == "" {
			failed = append(failed, types.ImportFailure{
				Channel: channelImport,
				Error:   "URL is required",
			})
			continue
		}

		// Extract channel ID from URL, with yt-dlp fallback for @username, /c/, /user/ URLs
		channelID, err := extractChannelIDWithYtdlpFallback(c.Request().Context(), h.ytdlpEnricher, channelImport.URL)
		if err != nil {
			failed = append(failed, types.ImportFailure{
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
				failed = append(failed, types.ImportFailure{
					Channel: channelImport,
					Error:   "Could not fetch channel title from RSS feed: " + err.Error(),
				})
				continue
			}
			title = feed.Title
			if title == "" {
				failed = append(failed, types.ImportFailure{
					Channel: channelImport,
					Error:   "Channel title could not be determined from RSS feed",
				})
				continue
			}
		}

		channel := store.Channel{ID: channelID, Title: title}
		if err := h.store.AddChannel(channel); err != nil {
			failed = append(failed, types.ImportFailure{
				Channel: channelImport,
				Error:   "Failed to add channel to database: " + err.Error(),
			})
			continue
		}

		storeChannel := store.Channel{ID: channelID, Title: title}
		imported = append(imported, types.TransformChannel(storeChannel))
	}

	response := types.ImportChannelsResponse{
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
