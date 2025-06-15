package ytdlp

import (
	"context"
	"fmt"

	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/videoid"
)

// MockEnricher is a mock implementation of Enricher for testing
type MockEnricher struct {
	ShouldFail bool
}

// NewMockEnricher creates a new mock enricher
func NewMockEnricher() *MockEnricher {
	return &MockEnricher{
		ShouldFail: false,
	}
}

// EnrichEntry enriches an RSS entry with mock yt-dlp data
func (m *MockEnricher) EnrichEntry(ctx context.Context, entry *rss.Entry) error {
	if m.ShouldFail {
		return fmt.Errorf("mock enricher configured to fail")
	}

	// Extract video ID for mock data using videoid package
	vid, err := videoid.NewFromFull(entry.ID)
	if err != nil {
		return fmt.Errorf("invalid entry ID format: %s - %w", entry.ID, err)
	}
	videoID := vid.ToRaw()

	// Add mock enhanced metadata
	entry.Duration = 300 // 5 minutes
	entry.Tags = []string{"technology", "tutorial", "programming", "demo"}
	entry.TopComments = []string{
		"Great video! Very informative.",
		"Thanks for the tutorial, this helped a lot.",
		"Could you make a follow-up video?",
		"Amazing content as always!",
		"This is exactly what I was looking for.",
	}
	entry.AutoSubtitles = "https://example.com/subtitles/" + videoID + ".vtt"

	return nil
}

// ResolveChannelID resolves a YouTube URL to a channel ID using mock data
func (m *MockEnricher) ResolveChannelID(ctx context.Context, url string) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("mock enricher configured to fail")
	}

	// Return mock channel IDs based on URL patterns for testing
	if contains := func(s, substr string) bool {
		return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
	}; contains(url, "@ChinaTalkMedi") {
		return "UCrAhw9Z8NI6GzO2WnvhYzCg", nil
	} else if contains(url, "@TestChannel") {
		return "UCTestChannelID123456789", nil
	}

	// Default mock channel ID for any other URL
	return "UCMockChannelID1234567890", nil
}
