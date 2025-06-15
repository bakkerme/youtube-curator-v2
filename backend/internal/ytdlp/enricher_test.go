package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/videoid"
)

// MockCommandExecutor is a mock implementation of CommandExecutor for testing
type MockCommandExecutor struct {
	ShouldFail    bool
	ShouldTimeout bool
	ReturnData    *YtdlpOutput
	Error         error
	ExecuteFunc   func(ctx context.Context, name string, args ...string) ([]byte, error)
}

// Execute implements CommandExecutor interface for testing
func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	// If a custom execute function is provided, use it
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, name, args...)
	}

	if m.ShouldTimeout {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	if m.ShouldFail {
		if m.Error != nil {
			return nil, m.Error
		}
		return nil, fmt.Errorf("mock command failed")
	}

	// Return mock JSON data
	if m.ReturnData != nil {
		return json.Marshal(m.ReturnData)
	}

	// Default mock data
	mockData := YtdlpOutput{
		Duration: 300,
		Tags:     []string{"technology", "tutorial"},
		Comments: []Comment{
			{Text: "Great video!", Author: "User1", LikeCount: 10},
			{Text: "Very helpful", Author: "User2", LikeCount: 5},
		},
		AutomaticCaptions: map[string][]SubtitleInfo{
			"en": {{Ext: "vtt", URL: "https://example.com/subs.vtt"}},
		},
	}

	return json.Marshal(mockData)
}

func TestEnrichEntry_InvalidVideoID(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}
	enricher := NewDefaultEnricherWithExecutor(mockExecutor)

	entry := &rss.Entry{
		ID: "invalid:format",
	}

	err := enricher.EnrichEntry(context.Background(), entry)
	if err == nil {
		t.Error("Expected error for invalid video ID format")
	}

	if !strings.Contains(err.Error(), "invalid entry ID format") {
		t.Errorf("Expected error about invalid entry ID format, got: %v", err)
	}
}

func TestEnrichEntry_Success(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		ReturnData: &YtdlpOutput{
			Duration: 420,
			Tags:     []string{"programming", "golang", "testing"},
			Comments: []Comment{
				{Text: "Excellent tutorial!", Author: "Developer1", LikeCount: 25},
				{Text: "Thanks for sharing", Author: "Developer2", LikeCount: 15},
				{Text: "Very clear explanation", Author: "Developer3", LikeCount: 8},
			},
			AutomaticCaptions: map[string][]SubtitleInfo{
				"en": {{Ext: "vtt", URL: "https://example.com/test-subs.vtt"}},
			},
		},
	}

	enricher := NewDefaultEnricherWithExecutor(mockExecutor)

	entry := &rss.Entry{
		ID: "yt:video:testVideoID",
	}

	err := enricher.EnrichEntry(context.Background(), entry)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify enrichment worked
	if entry.Duration != 420 {
		t.Errorf("Expected duration 420, got %d", entry.Duration)
	}

	expectedTags := []string{"programming", "golang", "testing"}
	if len(entry.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(entry.Tags))
	}

	if len(entry.TopComments) != 3 {
		t.Errorf("Expected 3 top comments, got %d", len(entry.TopComments))
	}

	if entry.AutoSubtitles != "https://example.com/test-subs.vtt" {
		t.Errorf("Expected auto subtitles URL, got %s", entry.AutoSubtitles)
	}
}

func TestEnrichEntry_CommandFailure(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		ShouldFail: true,
		Error:      fmt.Errorf("command execution failed"),
	}

	enricher := NewDefaultEnricherWithExecutor(mockExecutor)

	entry := &rss.Entry{
		ID: "yt:video:dQw4w9WgXcQ",
	}

	err := enricher.EnrichEntry(context.Background(), entry)
	if err == nil {
		t.Error("Expected error for command failure")
	}

	if !strings.Contains(err.Error(), "yt-dlp command failed") {
		t.Errorf("Expected yt-dlp command failed error, got: %v", err)
	}
}

func TestEnrichEntry_Timeout(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		ShouldTimeout: true,
	}

	enricher := &DefaultEnricher{
		ytdlpPath:  "yt-dlp",
		timeout:    100 * time.Millisecond,
		maxRetries: 0, // No retries for this test
		executor:   mockExecutor,
	}

	entry := &rss.Entry{
		ID: "yt:video:dQw4w9WgXcQ",
	}

	err := enricher.EnrichEntry(context.Background(), entry)
	if err == nil {
		t.Error("Expected error for timeout")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestEnrichEntry_RetryLogic(t *testing.T) {
	// Test that retry logic works with retryable errors
	callCount := 0
	mockExecutor := &MockCommandExecutor{}

	// Set custom execute function to fail first two times, succeed on third
	mockExecutor.ExecuteFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		callCount++
		if callCount <= 2 {
			return nil, fmt.Errorf("network timeout occurred") // retryable error
		}
		// On third call, return success with default mock data
		mockData := YtdlpOutput{
			Duration: 300,
			Tags:     []string{"technology", "tutorial"},
			Comments: []Comment{
				{Text: "Great video!", Author: "User1", LikeCount: 10},
			},
			AutomaticCaptions: map[string][]SubtitleInfo{
				"en": {{Ext: "vtt", URL: "https://example.com/subs.vtt"}},
			},
		}
		return json.Marshal(mockData)
	}

	enricher := &DefaultEnricher{
		ytdlpPath:  "yt-dlp",
		timeout:    100 * time.Millisecond,
		maxRetries: 2,
		executor:   mockExecutor,
	}

	entry := &rss.Entry{
		ID: "yt:video:dQw4w9WgXcQ",
	}

	err := enricher.EnrichEntry(context.Background(), entry)
	if err != nil {
		t.Fatalf("Expected success after retries, got: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 command executions (2 failures + 1 success), got %d", callCount)
	}

	// Verify enrichment worked
	if entry.Duration == 0 {
		t.Error("Expected duration to be set after successful retry")
	}
}

func TestEnrichEntry_NonRetryableError(t *testing.T) {
	callCount := 0
	mockExecutor := &MockCommandExecutor{}

	// Set custom execute function to always fail with non-retryable error
	mockExecutor.ExecuteFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		callCount++
		return nil, fmt.Errorf("invalid video format") // non-retryable error
	}

	enricher := &DefaultEnricher{
		ytdlpPath:  "yt-dlp",
		timeout:    1 * time.Second,
		maxRetries: 2,
		executor:   mockExecutor,
	}

	entry := &rss.Entry{
		ID: "yt:video:dQw4w9WgXcQ",
	}

	err := enricher.EnrichEntry(context.Background(), entry)
	if err == nil {
		t.Error("Expected error for non-retryable failure")
	}

	if callCount != 1 {
		t.Errorf("Expected 1 command execution (no retries for non-retryable error), got %d", callCount)
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

	if enricher.executor == nil {
		t.Error("Expected executor to be set")
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

	if enricher.executor == nil {
		t.Error("Expected executor to be set")
	}
}

func TestNewDefaultEnricherWithExecutor(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}
	enricher := NewDefaultEnricherWithExecutor(mockExecutor)

	if enricher.executor != mockExecutor {
		t.Error("Expected custom executor to be set")
	}

	if enricher.timeout != 60*time.Second {
		t.Errorf("Expected default timeout 60s, got %v", enricher.timeout)
	}

	if enricher.maxRetries != 2 {
		t.Errorf("Expected default maxRetries 2, got %v", enricher.maxRetries)
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
