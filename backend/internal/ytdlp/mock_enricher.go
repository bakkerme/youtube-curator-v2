package ytdlp

import (
	"context"
	"fmt"

	"youtube-curator-v2/internal/rss"
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

	// Extract video ID for mock data
	videoID, err := extractVideoID(entry.ID)
	if err != nil {
		return err
	}

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
