package handlers

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"youtube-curator-v2/internal/api/types"
	"youtube-curator-v2/internal/rss"

	"github.com/labstack/echo/v4"
)

// VideoHandlers provides handlers for video operations and summaries
type VideoHandlers struct {
	*BaseHandlers
}

// NewVideoHandlers creates a new instance of video handlers
func NewVideoHandlers(base *BaseHandlers) *VideoHandlers {
	return &VideoHandlers{BaseHandlers: base}
}

// GetVideos handles GET /api/videos - returns all videos from the video store
func (h *VideoHandlers) GetVideos(c echo.Context) error {
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

	// Prepare response using the transformation function
	response := types.TransformVideos(videos, h.videoStore.GetLastRefreshedAt())

	return c.JSON(http.StatusOK, response)
}

// MarkVideoAsWatched handles POST /api/videos/:videoId/watch
func (h *VideoHandlers) MarkVideoAsWatched(c echo.Context) error {
	rawVideoID := c.Param("videoId")
	if rawVideoID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Video ID is required")
	}

	// Convert raw video ID to the format expected by the system (yt:video:ID)
	videoID := "yt:video:" + rawVideoID

	// Validate video ID format using the dedicated validator
	if err := rss.ValidateYouTubeVideoID(videoID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if h.videoStore == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Video store not initialized")
	}

	// Call the store function to mark the video as watched.
	// Note: The current videoStore.MarkVideoAsWatched doesn't return an error or status
	// if the video is not found. It simply does nothing in that case.
	// For a more robust API, the store method could be enhanced to return a boolean or error.
	h.videoStore.MarkVideoAsWatched(videoID)

	// Since the store method doesn't indicate if the video was found,
	// we will assume success if no other errors occurred.
	// A better approach would be for MarkVideoAsWatched to return a status.
	return c.NoContent(http.StatusNoContent) // HTTP 204 No Content is suitable for successful actions with no response body
}

// GetVideoSummary handles GET /api/videos/:videoId/summary
func (h *VideoHandlers) GetVideoSummary(c echo.Context) error {
	rawVideoID := c.Param("videoId")
	if rawVideoID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Video ID is required")
	}

	// Convert raw video ID to the format expected by the system (yt:video:ID)
	videoID := "yt:video:" + rawVideoID

	// Validate video ID format using the dedicated validator
	if err := rss.ValidateYouTubeVideoID(videoID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Check if summary service is available
	if h.summaryService == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Summary service not available")
	}

	// Get or generate summary
	ctx := c.Request().Context()
	result := h.summaryService.GetOrGenerateSummary(ctx, videoID)

	if result.Error != nil {
		// Handle different types of errors
		switch {
		case strings.Contains(result.Error.Error(), "LLM not configured"):
			return echo.NewHTTPError(http.StatusServiceUnavailable, "LLM service not configured")
		case strings.Contains(result.Error.Error(), "no subtitles available"):
			return echo.NewHTTPError(http.StatusNotFound, "No subtitles available for this video")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, result.Error.Error())
		}
	}

	response := types.VideoSummaryResponse{
		VideoID:        videoID,
		Summary:        result.Summary,
		Thinking:       result.Thinking,
		SourceLanguage: result.SourceLanguage,
		GeneratedAt:    result.GeneratedAt.Format(time.RFC3339),
		Tracked:        result.Tracked,
	}

	return c.JSON(http.StatusOK, response)
}