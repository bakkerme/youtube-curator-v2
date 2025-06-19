package summary

import (
	"context"
	"testing"
	"time"

	"youtube-curator-v2/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockService_GetOrGenerateSummary(t *testing.T) {
	// Create a mock store (we could use gomock here in a real implementation)
	mockStore := &mockStore{}
	service := NewMockService(mockStore)

	tests := []struct {
		name     string
		videoID  string
		expected string
	}{
		{
			name:     "tech video",
			videoID:  "tech123456",
			expected: "technical concepts",
		},
		{
			name:     "tutorial video",
			videoID:  "tutorial789",
			expected: "educational tutorial",
		},
		{
			name:     "review video",
			videoID:  "review456",
			expected: "comprehensive review",
		},
		{
			name:     "default video",
			videoID:  "random123",
			expected: "interesting content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := service.GetOrGenerateSummary(ctx, tt.videoID)

			require.NotNil(t, result)
			assert.NoError(t, result.Error)
			assert.Equal(t, tt.videoID, result.VideoID)
			assert.Contains(t, result.Summary, tt.expected)
			assert.Equal(t, "en", result.SourceLanguage)
			assert.False(t, result.Tracked)
			assert.WithinDuration(t, time.Now(), result.GeneratedAt, time.Minute)
		})
	}
}

func TestMockService_VideoIDLength(t *testing.T) {
	mockStore := &mockStore{}
	service := NewMockService(mockStore)

	// Test with short video ID
	shortVideoID := "abc"
	ctx := context.Background()
	result := service.GetOrGenerateSummary(ctx, shortVideoID)

	require.NotNil(t, result)
	assert.NoError(t, result.Error)
	// Should handle short video IDs gracefully
	assert.Contains(t, result.Summary, "abc")
}

// mockStore implements the basic Store interface for testing
type mockStore struct{}

func (m *mockStore) Close() error                                           { return nil }
func (m *mockStore) GetLastCheckedVideoID(channelID string) (string, error) { return "", nil }
func (m *mockStore) SetLastCheckedVideoID(channelID, videoID string) error  { return nil }
func (m *mockStore) GetLastCheckedTimestamp(channelID string) (time.Time, error) {
	return time.Time{}, nil
}
func (m *mockStore) SetLastCheckedTimestamp(channelID string, timestamp time.Time) error { return nil }
func (m *mockStore) GetChannels() ([]store.Channel, error)                               { return nil, nil }
func (m *mockStore) AddChannel(channel store.Channel) error                              { return nil }
func (m *mockStore) RemoveChannel(channelID string) error                                { return nil }
func (m *mockStore) GetCheckInterval() (time.Duration, error)                            { return time.Hour, nil }
func (m *mockStore) SetCheckInterval(interval time.Duration) error                       { return nil }
func (m *mockStore) GetSMTPConfig() (*store.SMTPConfig, error)                           { return nil, nil }
func (m *mockStore) SetSMTPConfig(config *store.SMTPConfig) error                        { return nil }
func (m *mockStore) GetLLMConfig() (*store.LLMConfig, error)                             { return nil, nil }
func (m *mockStore) SetLLMConfig(config *store.LLMConfig) error                          { return nil }
func (m *mockStore) GetNewsletterConfig() (*store.NewsletterConfig, error)              { return nil, nil }
func (m *mockStore) SetNewsletterConfig(config *store.NewsletterConfig) error           { return nil }
func (m *mockStore) GetWatchedVideos() ([]string, error)                           { return nil, nil }
func (m *mockStore) SetVideoWatched(videoID string) error                           { return nil }
func (m *mockStore) IsVideoWatched(videoID string) (bool, error)                    { return false, nil }
