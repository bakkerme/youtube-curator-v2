package ytdlp

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"youtube-curator-v2/internal/rss"
)

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		name     string
		entryID  string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid youtube entry ID",
			entryID:  "yt:video:dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
			wantErr:  false,
		},
		{
			name:     "invalid format - too few parts",
			entryID:  "yt:video",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid format - wrong prefix",
			entryID:  "youtube:video:dQw4w9WgXcQ",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid format - wrong type",
			entryID:  "yt:channel:dQw4w9WgXcQ",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractVideoID(tt.entryID)

			if tt.wantErr && err == nil {
				t.Errorf("extractVideoID() expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("extractVideoID() unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("extractVideoID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnrichEntry_InvalidVideoID(t *testing.T) {
	enricher := NewDefaultEnricher()

	entry := &rss.Entry{
		ID: "invalid:format",
	}

	err := enricher.EnrichEntry(context.Background(), entry)
	if err == nil {
		t.Error("Expected error for invalid video ID format")
	}

	if !strings.Contains(err.Error(), "failed to extract video ID") {
		t.Errorf("Expected error about video ID extraction, got: %v", err)
	}
}

func TestEnrichEntry_Timeout(t *testing.T) {
	// Create enricher with very short timeout and non-existent command
	enricher := &DefaultEnricher{
		ytdlpPath: "nonexistent-command", // Command that doesn't exist to trigger error
		timeout:   100 * time.Millisecond,
	}

	entry := &rss.Entry{
		ID: "yt:video:dQw4w9WgXcQ",
	}

	err := enricher.EnrichEntry(context.Background(), entry)
	if err == nil {
		t.Error("Expected error for non-existent command")
	}

	// Should get an error about the command failing
	if !strings.Contains(err.Error(), "yt-dlp command failed") {
		t.Errorf("Expected yt-dlp command failed error, got: %v", err)
	}
}

func TestEnrichEntry_RetryLogic(t *testing.T) {
	// Test that retry logic is properly configured
	enricher := NewDefaultEnricher()

	if enricher.maxRetries != 2 {
		t.Errorf("Expected maxRetries 2, got %v", enricher.maxRetries)
	}

	// Test isRetryableError function with network-related errors
	networkErr := fmt.Errorf("network timeout occurred")
	if !isRetryableError(networkErr) {
		t.Error("Expected network timeout to be retryable")
	}

	// Test that non-retryable errors aren't retried
	invalidErr := fmt.Errorf("invalid video format")
	if isRetryableError(invalidErr) {
		t.Error("Expected invalid format error to not be retryable")
	}
}

func TestNewDefaultEnricherWithTimeout(t *testing.T) {
	timeout := 45 * time.Second
	enricher := NewDefaultEnricherWithTimeout(timeout)

	if enricher.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, enricher.timeout)
	}

	if enricher.maxRetries != 2 {
		t.Errorf("Expected maxRetries 2, got %v", enricher.maxRetries)
	}
}

func TestNewDefaultEnricherWithConfig(t *testing.T) {
	timeout := 45 * time.Second
	maxRetries := 5
	enricher := NewDefaultEnricherWithConfig(timeout, maxRetries)

	if enricher.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, enricher.timeout)
	}

	if enricher.maxRetries != maxRetries {
		t.Errorf("Expected maxRetries %v, got %v", maxRetries, enricher.maxRetries)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "timeout error",
			err:      fmt.Errorf("connection timeout"),
			expected: true,
		},
		{
			name:     "network error",
			err:      fmt.Errorf("network unreachable"),
			expected: true,
		},
		{
			name:     "connection error",
			err:      fmt.Errorf("connection refused"),
			expected: true,
		},
		{
			name:     "temporary failure",
			err:      fmt.Errorf("temporary failure in name resolution"),
			expected: true,
		},
		{
			name:     "permanent error",
			err:      fmt.Errorf("invalid video ID"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}
