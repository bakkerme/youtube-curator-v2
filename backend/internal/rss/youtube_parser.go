package rss

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ExtractChannelID extracts a YouTube channel ID from various URL formats
// Supports:
// - https://www.youtube.com/channel/CHANNEL_ID
// - https://www.youtube.com/@USERNAME
// - https://www.youtube.com/c/CUSTOM_NAME
// - https://www.youtube.com/user/USERNAME
// - Direct channel ID input
func ExtractChannelID(input string) (string, error) {
	// First, check if input is already a valid channel ID (starts with UC and is 24 characters)
	if isValidChannelID(input) {
		return input, nil
	}

	// Parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid URL or channel ID: %w", err)
	}

	// Handle different YouTube URL formats
	path := strings.TrimPrefix(parsedURL.Path, "/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) < 2 {
		return "", fmt.Errorf("invalid YouTube URL format")
	}

	switch pathParts[0] {
	case "channel":
		// Direct channel URL: https://www.youtube.com/channel/CHANNEL_ID
		channelID := pathParts[1]
		if !isValidChannelID(channelID) {
			return "", fmt.Errorf("invalid channel ID format: %s", channelID)
		}
		return channelID, nil
	case "c", "user":
		// Custom URL or user URL - these cannot be resolved without the YouTube API
		// Return an error asking the user to provide the channel ID directly
		return "", fmt.Errorf("custom URLs (@username, /c/, /user/) are not supported. Please provide the channel ID directly (starts with 'UC')")
	default:
		// Handle @username format
		if strings.HasPrefix(pathParts[0], "@") {
			return "", fmt.Errorf("@username URLs are not supported. Please provide the channel ID directly (starts with 'UC')")
		}
		return "", fmt.Errorf("unsupported YouTube URL format")
	}
}

// isValidChannelID checks if a string is a valid YouTube channel ID
// YouTube channel IDs are typically 24 characters long and start with 'UC'
func isValidChannelID(id string) bool {
	// Channel ID regex pattern: starts with UC, followed by 22 alphanumeric characters and underscores/hyphens
	pattern := `^UC[a-zA-Z0-9_-]{22}$`
	matched, err := regexp.MatchString(pattern, id)
	return err == nil && matched
}

// ValidateChannelID validates that a channel ID is in the correct format
func ValidateChannelID(channelID string) error {
	if !isValidChannelID(channelID) {
		return fmt.Errorf("invalid channel ID format. Channel IDs should start with 'UC' and be 24 characters long")
	}
	return nil
}

const (
	youtubeVideoIDPrefix = "yt:video:"
	youtubeVideoIDPattern = `^[a-zA-Z0-9_-]{11}$`
)

var youtubeVideoIDRegexp = regexp.MustCompile(youtubeVideoIDPattern)

// ValidateYouTubeVideoID checks if a string is a valid YouTube video ID with the "yt:video:" prefix.
// The actual ID part must be 11 characters long and consist of alphanumeric characters, underscores, and hyphens.
func ValidateYouTubeVideoID(videoID string) error {
	if !strings.HasPrefix(videoID, youtubeVideoIDPrefix) {
		return fmt.Errorf("invalid video ID format. Expected prefix '%s'", youtubeVideoIDPrefix)
	}

	actualID := strings.TrimPrefix(videoID, youtubeVideoIDPrefix)

	if !youtubeVideoIDRegexp.MatchString(actualID) {
		return fmt.Errorf("invalid video ID format. Expected format: %s<11_alphanumeric_chars_hyphens_underscores>", youtubeVideoIDPrefix)
	}

	return nil
}
