package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"youtube-curator-v2/internal/api/types"
	"youtube-curator-v2/internal/videoid"

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
				if err := h.videoStore.AddVideo(channel.ID, mostRecentVideo); err != nil {
					log.Printf("Failed to add video %s to store: %v", mostRecentVideo.ID, err)
					// Continue processing other channels despite this error
				}
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

	// Create and validate video ID
	vid, err := videoid.NewFromRaw(rawVideoID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	videoID := vid.ToFull()

	if h.videoStore == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Video store not initialized")
	}

	// Call the store function to mark the video as watched.
	if err := h.videoStore.MarkVideoAsWatched(videoID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to mark video as watched: %v", err))
	}

	return c.NoContent(http.StatusNoContent) // HTTP 204 No Content is suitable for successful actions with no response body
}

// SetVideoToWatch handles POST /api/videos/:videoId/towatch
func (h *VideoHandlers) SetVideoToWatch(c echo.Context) error {
	rawVideoID := c.Param("videoId")
	if rawVideoID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Video ID is required")
	}

	// Create and validate video ID
	vid, err := videoid.NewFromRaw(rawVideoID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	videoID := vid.ToFull()

	if h.videoStore == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Video store not initialized")
	}

	// Call the store function to mark the video as to-watch
	if err := h.videoStore.SetVideoToWatch(videoID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to set video to watch: %v", err))
	}

	return c.NoContent(http.StatusNoContent)
}

// UnsetVideoToWatch handles DELETE /api/videos/:videoId/towatch
func (h *VideoHandlers) UnsetVideoToWatch(c echo.Context) error {
	rawVideoID := c.Param("videoId")
	if rawVideoID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Video ID is required")
	}

	// Create and validate video ID
	vid, err := videoid.NewFromRaw(rawVideoID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	videoID := vid.ToFull()

	if h.videoStore == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Video store not initialized")
	}

	// Call the store function to unset the video from to-watch
	if err := h.videoStore.UnsetVideoToWatch(videoID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to unset video to watch: %v", err))
	}

	return c.NoContent(http.StatusNoContent)
}

// GetVideoSummary handles GET /api/videos/:videoId/summary
func (h *VideoHandlers) GetVideoSummary(c echo.Context) error {
	rawVideoID := c.Param("videoId")
	if rawVideoID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Video ID is required")
	}

	// Create and validate video ID
	vid, err := videoid.NewFromRaw(rawVideoID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	videoID := vid.ToFull()

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