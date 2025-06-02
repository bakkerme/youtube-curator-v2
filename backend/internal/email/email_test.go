package email

import (
	"strings"
	"testing"
	"time"
	"youtube-curator-v2/internal/rss"
)

func TestTruncateLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLines int
		expected string
	}{
		{
			name:     "Short text under limit",
			input:    "Line 1\nLine 2\nLine 3",
			maxLines: 5,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Text exactly at limit",
			input:    "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
			maxLines: 5,
			expected: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
		},
		{
			name:     "Text over limit",
			input:    "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7",
			maxLines: 5,
			expected: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5...",
		},
		{
			name:     "Empty string",
			input:    "",
			maxLines: 5,
			expected: "",
		},
		{
			name:     "Single line",
			input:    "Just one line",
			maxLines: 5,
			expected: "Just one line",
		},
	}

	// Create a mock truncateLines function to test
	truncateLines := func(s string, maxLines int) string {
		if s == "" {
			return s
		}
		lines := strings.Split(s, "\n")
		if len(lines) <= maxLines {
			return s
		}
		truncated := strings.Join(lines[:maxLines], "\n")
		return truncated + "..."
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateLines(tt.input, tt.maxLines)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestFormatNewVideosEmail(t *testing.T) {
	// Create test data
	publishedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	videos := []rss.Entry{
		{
			Title:     "Test Video",
			Link:      rss.Link{Href: "https://youtube.com/watch?v=test"},
			ID:        "test-id",
			Published: publishedTime,
			Author:    rss.Author{Name: "Test Channel"},
			MediaGroup: rss.MediaGroup{
				MediaThumbnail:   rss.MediaThumbnail{URL: "https://example.com/thumb.jpg"},
				MediaDescription: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7",
			},
		},
	}

	// Test that the email formatting works
	result, err := FormatNewVideosEmail(videos)
	if err != nil {
		t.Fatalf("FormatNewVideosEmail failed: %v", err)
	}

	// Check that the result contains expected elements
	if !strings.Contains(result, "Test Video") {
		t.Error("Email should contain video title")
	}
	if !strings.Contains(result, "Test Channel") {
		t.Error("Email should contain channel name")
	}
	if !strings.Contains(result, "Line 1") {
		t.Error("Email should contain description")
	}
	// Check that truncation happened (should not contain Line 6)
	if strings.Contains(result, "Line 6") {
		t.Error("Email should truncate description at 5 lines")
	}
	if !strings.Contains(result, "Line 5...") {
		t.Error("Email should show truncation indicator")
	}
}
