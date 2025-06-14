package videoid

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// YouTubeVideoIDPrefix is the prefix used in YouTube RSS feeds
	YouTubeVideoIDPrefix = "yt:video:"
	// YouTubeVideoIDLength is the expected length of a raw YouTube video ID
	YouTubeVideoIDLength = 11
	// YouTubeVideoIDPattern is the regex pattern for valid YouTube video IDs
	YouTubeVideoIDPattern = `^[a-zA-Z0-9_-]{11}$`
)

var youtubeVideoIDRegexp = regexp.MustCompile(YouTubeVideoIDPattern)

// VideoID represents a YouTube video ID that can be in either full or raw format
type VideoID struct {
	raw string
}

// NewFromRaw creates a VideoID from a raw video ID (11 characters)
func NewFromRaw(raw string) (*VideoID, error) {
	if err := validateRawVideoID(raw); err != nil {
		return nil, err
	}
	return &VideoID{raw: raw}, nil
}

// NewFromFull creates a VideoID from a full video ID (yt:video:ABC123)
func NewFromFull(full string) (*VideoID, error) {
	if !strings.HasPrefix(full, YouTubeVideoIDPrefix) {
		return nil, fmt.Errorf("invalid full video ID format: missing prefix %s", YouTubeVideoIDPrefix)
	}

	raw := strings.TrimPrefix(full, YouTubeVideoIDPrefix)
	if err := validateRawVideoID(raw); err != nil {
		return nil, err
	}

	return &VideoID{raw: raw}, nil
}

// ToRaw returns the raw video ID (11 characters)
func (v *VideoID) ToRaw() string {
	return v.raw
}

// ToFull returns the full video ID with yt:video: prefix
func (v *VideoID) ToFull() string {
	return YouTubeVideoIDPrefix + v.raw
}

// String returns the full video ID format for compatibility
func (v *VideoID) String() string {
	return v.ToFull()
}

// validateRawVideoID validates that a raw video ID is 11 characters and matches the expected pattern
func validateRawVideoID(raw string) error {
	if len(raw) != YouTubeVideoIDLength {
		return fmt.Errorf("invalid raw video ID length: expected %d characters, got %d", YouTubeVideoIDLength, len(raw))
	}

	if !youtubeVideoIDRegexp.MatchString(raw) {
		return fmt.Errorf("invalid raw video ID format: must match pattern %s", YouTubeVideoIDPattern)
	}

	return nil
}

// ValidateFullVideoID validates a full video ID format
func ValidateFullVideoID(full string) error {
	_, err := NewFromFull(full)
	return err
}

// ValidateRawVideoID validates a raw video ID format
func ValidateRawVideoID(raw string) error {
	_, err := NewFromRaw(raw)
	return err
}
