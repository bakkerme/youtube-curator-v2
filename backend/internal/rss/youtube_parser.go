package rss

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"youtube-curator-v2/internal/videoid"
)

// ChannelIDResolver interface for resolving YouTube URLs to channel IDs
type ChannelIDResolver interface {
	ResolveChannelID(ctx context.Context, url string) (string, error)
}

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
		return "", fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	// Handle different YouTube URL formats
	path := strings.TrimPrefix(parsedURL.Path, "/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) < 2 {
		return "", ErrInvalidURLFormat
	}

	switch pathParts[0] {
	case "channel":
		// Direct channel URL: https://www.youtube.com/channel/CHANNEL_ID
		channelID := pathParts[1]
		if !isValidChannelID(channelID) {
			return "", NewInvalidChannelIDError(channelID)
		}
		return channelID, nil
	case "c", "user":
		// Custom URL or user URL - these cannot be resolved without yt-dlp
		return "", NewResolverRequiredError("custom URLs (/c/, /user/)")
	default:
		// Handle @username format
		if strings.HasPrefix(pathParts[0], "@") {
			return "", NewResolverRequiredError("@username")
		}
		return "", ErrUnsupportedURLFormat
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
		return NewInvalidChannelIDError(channelID)
	}
	return nil
}

const (
	youtubeVideoIDPrefix  = "yt:video:"
	youtubeVideoIDPattern = `^[a-zA-Z0-9_-]{11}$`
)

var youtubeVideoIDRegexp = regexp.MustCompile(youtubeVideoIDPattern)

// ValidateYouTubeVideoID checks if a string is a valid YouTube video ID with the "yt:video:" prefix.
// The actual ID part must be 11 characters long and consist of alphanumeric characters, underscores, and hyphens.
// Deprecated: Use videoid.ValidateFullVideoID instead
func ValidateYouTubeVideoID(videoID string) error {
	err := videoid.ValidateFullVideoID(videoID)
	if err != nil {
		// Return the original ValidationError type for backward compatibility
		return NewInvalidVideoIDError()
	}
	return nil
}

// ExtractChannelIDWithResolver extracts a YouTube channel ID from various URL formats using an optional resolver
// This function can handle @username, /c/, and /user/ URLs when a resolver is provided
func ExtractChannelIDWithResolver(ctx context.Context, input string, resolver ChannelIDResolver) (string, error) {
	// First, check if input is already a valid channel ID (starts with UC and is 24 characters)
	if isValidChannelID(input) {
		return input, nil
	}

	// Parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	// Handle different YouTube URL formats
	path := strings.TrimPrefix(parsedURL.Path, "/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) < 1 {
		return "", ErrInvalidURLFormat
	}

	switch pathParts[0] {
	case "channel":
		// Direct channel URL: https://www.youtube.com/channel/CHANNEL_ID
		if len(pathParts) < 2 {
			return "", ErrInvalidURLFormat
		}
		channelID := pathParts[1]
		if !isValidChannelID(channelID) {
			return "", NewInvalidChannelIDError(channelID)
		}
		return channelID, nil
	case "c", "user":
		// Custom URL or user URL - use resolver if provided
		if resolver == nil {
			return "", NewResolverRequiredError("custom URLs (/c/, /user/)")
		}
		return resolveWithFallback(ctx, input, resolver)
	default:
		// Handle @username format
		if strings.HasPrefix(pathParts[0], "@") {
			if resolver == nil {
				return "", NewResolverRequiredError("@username")
			}
			return resolveWithFallback(ctx, input, resolver)
		}
		return "", ErrUnsupportedURLFormat
	}
}

// resolveWithFallback uses the resolver to get a channel ID and validates it
func resolveWithFallback(ctx context.Context, url string, resolver ChannelIDResolver) (string, error) {
	// Use the resolver to get the channel ID
	channelID, err := resolver.ResolveChannelID(ctx, url)
	if err != nil {
		return "", fmt.Errorf("%w for URL %s: %w", ErrResolverFailed, url, err)
	}

	// Validate the resolved channel ID
	if !isValidChannelID(channelID) {
		return "", fmt.Errorf("%w: %s", ErrResolvedIDInvalid, channelID)
	}

	return channelID, nil
}
