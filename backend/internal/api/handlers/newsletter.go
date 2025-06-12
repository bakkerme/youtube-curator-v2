package handlers

import (
	"net/http"
	"sort"

	"youtube-curator-v2/internal/api/types"
	"youtube-curator-v2/internal/email"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"

	"github.com/labstack/echo/v4"
)

// NewsletterHandlers provides handlers for newsletter manual triggers
type NewsletterHandlers struct {
	*BaseHandlers
}

// NewNewsletterHandlers creates a new instance of newsletter handlers
func NewNewsletterHandlers(base *BaseHandlers) *NewsletterHandlers {
	return &NewsletterHandlers{BaseHandlers: base}
}

// RunNewsletter handles POST /api/newsletter/run
func (h *NewsletterHandlers) RunNewsletter(c echo.Context) error {
	var req types.RunNewsletterRequest
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

	ctx := c.Request().Context()
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
	response := types.NewsletterRunResponse{
		Message:           "Newsletter run completed",
		ChannelsProcessed: processedCount,
		ChannelsWithError: errorCount,
		NewVideosFound:    len(allNewVideos),
		EmailSent:         len(allNewVideos) > 0,
	}

	return c.JSON(http.StatusOK, response)
}
